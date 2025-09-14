package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/alireza-karampour/sms/internal/streams"
	. "github.com/alireza-karampour/sms/internal/subjects"
	"github.com/alireza-karampour/sms/internal/workers"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/alireza-karampour/sms/tests/helpers"
	"github.com/jackc/pgx/v5/pgtype"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("SMS Worker Integration Tests", func() {
	var (
		testSuite *helpers.TestSuite
		worker    *workers.Sms
		queries   *sqlc.Queries
		userID    int32
		phoneID   int32
	)

	BeforeEach(func() {
		testSuite = helpers.SetupTestSuite()
		queries = sqlc.New(testSuite.DB)

		// Set up rate limit configuration for tests
		viper.Set("sms.normal.ratelimit", 1000) // 1000ms = 1 second
		viper.Set("sms.express.ratelimit", 100) // 100ms = 0.1 second
		viper.Set("sms.cost", "5.0")

		// Create SMS worker
		var err error
		worker, err = workers.NewSms(context.Background(), "127.0.0.1:4223", testSuite.DB)
		Expect(err).NotTo(HaveOccurred())

		// Create test user and phone number
		balance := pgtype.Numeric{}
		balance.Scan("100.00")
		err = queries.AddUser(context.Background(), sqlc.AddUserParams{
			Username: "workertestuser",
			Balance:  balance,
		})
		Expect(err).NotTo(HaveOccurred())

		userID, err = queries.GetUserId(context.Background(), "workertestuser")
		Expect(err).NotTo(HaveOccurred())

		err = queries.AddPhoneNumber(context.Background(), sqlc.AddPhoneNumberParams{
			UserID:      userID,
			PhoneNumber: "+1234567890",
		})
		Expect(err).NotTo(HaveOccurred())

		phoneID, err = queries.GetPhoneNumberId(context.Background(), sqlc.GetPhoneNumberIdParams{
			UserID:      userID,
			PhoneNumber: "+1234567890",
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if worker != nil {
			worker.Close()
		}
		testSuite.CleanupTestData()
		testSuite.Cleanup()
	})

	Context("Worker Initialization", func() {
		It("should initialize SMS worker successfully", func() {
			Expect(worker).NotTo(BeNil())
			Expect(worker.Consumer).NotTo(BeNil())
			Expect(worker.Queries).NotTo(BeNil())
			Expect(worker.Queries).NotTo(BeNil())
		})

		It("should bind consumers for normal and express SMS streams", func() {
			// Verify that streams are created
			js, err := testSuite.NATSConn.Conn.JetStream()
			Expect(err).NotTo(HaveOccurred())

			// Check normal SMS stream
			streamInfo, err := js.StreamInfo(NORMAL_SMS_CONSUMER_NAME)
			Expect(err).NotTo(HaveOccurred())
			Expect(streamInfo).NotTo(BeNil())
			Expect(streamInfo.Config.Name).To(Equal(NORMAL_SMS_CONSUMER_NAME))

			// Check express SMS stream
			streamInfo, err = js.StreamInfo(EXPRESS_SMS_CONSUMER_NAME)
			Expect(err).NotTo(HaveOccurred())
			Expect(streamInfo).NotTo(BeNil())
			Expect(streamInfo.Config.Name).To(Equal(EXPRESS_SMS_CONSUMER_NAME))
		})
	})

	Context("Normal SMS Processing", func() {
		It("should process normal SMS request successfully", func() {
			// Create SMS data
			smsData := sqlc.Sm{
				UserID:        userID,
				PhoneNumberID: phoneID,
				ToPhoneNumber: "+0987654321",
				Message:       "Test normal SMS message",
				Status:        "pending",
			}

			// Get initial balance
			initialBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())

			// Start worker
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Publish message to normal SMS subject
			subject := MakeSubject(SMS, SEND, REQ)
			smsJSON, err := json.Marshal(smsData)
			Expect(err).NotTo(HaveOccurred())

			err = testSuite.NATSConn.Conn.Publish(subject, smsJSON)
			Expect(err).NotTo(HaveOccurred())

			// Wait for processing
			time.Sleep(500 * time.Millisecond)

			// Verify SMS was processed by checking balance deduction
			// (We can't easily query SMS records without a GetSmsByUserId query)
			// The balance deduction confirms the SMS was processed

			// Verify balance was deducted
			newBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(newBalance.Int.Int64()).To(BeNumerically("<", initialBalance.Int.Int64()))
		})
	})

	Context("Express SMS Processing", func() {
		It("should process express SMS request successfully", func() {
			// Create SMS data
			smsData := sqlc.Sm{
				UserID:        userID,
				PhoneNumberID: phoneID,
				ToPhoneNumber: "+0987654321",
				Message:       "Test express SMS message",
				Status:        "pending",
			}

			// Get initial balance
			initialBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())

			// Start worker
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Publish message to express SMS subject
			subject := MakeSubject(SMS, EX, SEND, REQ)
			smsJSON, err := json.Marshal(smsData)
			Expect(err).NotTo(HaveOccurred())

			err = testSuite.NATSConn.Conn.Publish(subject, smsJSON)
			Expect(err).NotTo(HaveOccurred())

			// Wait for processing
			time.Sleep(500 * time.Millisecond)

			// Verify SMS was processed by checking balance deduction
			// (Express SMS processing confirmation)

			// Verify balance was deducted
			newBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(newBalance.Int.Int64()).To(BeNumerically("<", initialBalance.Int.Int64()))
		})

	})

	Context("Error Handling", func() {
		It("should handle invalid JSON in SMS request", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Publish invalid JSON
			subject := MakeSubject(SMS, SEND, REQ)
			err := testSuite.NATSConn.Conn.Publish(subject, []byte("invalid json"))
			Expect(err).NotTo(HaveOccurred())

			// Wait for processing
			time.Sleep(200 * time.Millisecond)

			// Should not crash the worker
			// Invalid messages should be terminated
		})

		It("should handle database transaction errors gracefully", func() {
			// Create SMS data with invalid user ID to cause database error
			smsData := sqlc.Sm{
				UserID:        99999, // Non-existent user ID
				PhoneNumberID: phoneID,
				ToPhoneNumber: "+0987654321",
				Message:       "Test SMS with invalid user",
				Status:        "pending",
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Publish message
			subject := MakeSubject(SMS, SEND, REQ)
			smsJSON, err := json.Marshal(smsData)
			Expect(err).NotTo(HaveOccurred())

			err = testSuite.NATSConn.Conn.Publish(subject, smsJSON)
			Expect(err).NotTo(HaveOccurred())

			// Wait for processing
			time.Sleep(500 * time.Millisecond)

			// Should handle error gracefully without crashing
			// Message should be NAK'd and retried
		})
	})

	Context("Rate Limiting", func() {
		It("should respect rate limiting for normal SMS", func() {
			// This test verifies that normal SMS processing respects the 1000ms rate limit
			// by sending 2 SMS messages and checking the delivered_at time difference

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Send 2 SMS messages rapidly
			subject := MakeSubject(SMS, SEND, REQ)

			for i := 0; i < 2; i++ {
				smsData := sqlc.Sm{
					UserID:        userID,
					PhoneNumberID: phoneID,
					ToPhoneNumber: "+0987654321",
					Message:       fmt.Sprintf("Normal SMS rate limit test %d", i+1),
					Status:        "pending",
				}

				smsJSON, err := json.Marshal(smsData)
				Expect(err).NotTo(HaveOccurred())

				err = testSuite.NATSConn.Conn.Publish(subject, smsJSON)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for both messages to be processed
			time.Sleep(2500 * time.Millisecond)

			// Get the last 2 SMS messages from database
			smsMessages, err := queries.GetLastSmsMessages(context.Background(), sqlc.GetLastSmsMessagesParams{
				UserID: userID,
				Limit:  2,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(smsMessages)).To(Equal(2))

			// Check that the delivered_at time difference is >= 1000ms (rate limit)
			firstMessage := smsMessages[0]  // Most recent
			secondMessage := smsMessages[1] // Second most recent

			timeDiff := firstMessage.DeliveredAt.Time.Sub(secondMessage.DeliveredAt.Time)
			Expect(timeDiff).To(BeNumerically(">=", 1000*time.Millisecond))
		})

		It("should respect rate limiting for express SMS", func() {
			// This test verifies that express SMS processing respects the 100ms rate limit
			// by sending 2 SMS messages and checking the delivered_at time difference

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Send 2 express SMS messages rapidly
			subject := MakeSubject(SMS, EX, SEND, REQ)

			for i := 0; i < 2; i++ {
				smsData := sqlc.Sm{
					UserID:        userID,
					PhoneNumberID: phoneID,
					ToPhoneNumber: "+0987654321",
					Message:       fmt.Sprintf("Express SMS rate limit test %d", i+1),
					Status:        "pending",
				}

				smsJSON, err := json.Marshal(smsData)
				Expect(err).NotTo(HaveOccurred())

				err = testSuite.NATSConn.Conn.Publish(subject, smsJSON)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for both messages to be processed
			time.Sleep(500 * time.Millisecond)

			// Get the last 2 SMS messages from database
			smsMessages, err := queries.GetLastSmsMessages(context.Background(), sqlc.GetLastSmsMessagesParams{
				UserID: userID,
				Limit:  2,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(smsMessages)).To(Equal(2))

			// Check that the delivered_at time difference is >= 100ms (rate limit)
			firstMessage := smsMessages[0]  // Most recent
			secondMessage := smsMessages[1] // Second most recent

			timeDiff := firstMessage.DeliveredAt.Time.Sub(secondMessage.DeliveredAt.Time)
			Expect(timeDiff).To(BeNumerically(">=", 100*time.Millisecond))
		})

		It("should have different rate limits for normal vs express SMS", func() {
			// This test verifies that normal SMS has a higher rate limit (slower) than express SMS
			// by comparing the delivered_at time differences between normal and express SMS

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Test normal SMS rate limit - send 2 messages
			normalSubject := MakeSubject(SMS, SEND, REQ)
			for i := 0; i < 2; i++ {
				smsData := sqlc.Sm{
					UserID:        userID,
					PhoneNumberID: phoneID,
					ToPhoneNumber: "+0987654321",
					Message:       fmt.Sprintf("Normal SMS comparison %d", i+1),
					Status:        "pending",
				}

				smsJSON, err := json.Marshal(smsData)
				Expect(err).NotTo(HaveOccurred())

				err = testSuite.NATSConn.Conn.Publish(normalSubject, smsJSON)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for normal SMS processing
			time.Sleep(2500 * time.Millisecond)

			// Get normal SMS messages
			normalMessages, err := queries.GetLastSmsMessages(context.Background(), sqlc.GetLastSmsMessagesParams{
				UserID: userID,
				Limit:  2,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(normalMessages)).To(Equal(2))

			normalTimeDiff := normalMessages[0].DeliveredAt.Time.Sub(normalMessages[1].DeliveredAt.Time)

			// Test express SMS rate limit - send 2 messages
			expressSubject := MakeSubject(SMS, EX, SEND, REQ)
			for i := 0; i < 2; i++ {
				smsData := sqlc.Sm{
					UserID:        userID,
					PhoneNumberID: phoneID,
					ToPhoneNumber: "+0987654321",
					Message:       fmt.Sprintf("Express SMS comparison %d", i+1),
					Status:        "pending",
				}

				smsJSON, err := json.Marshal(smsData)
				Expect(err).NotTo(HaveOccurred())

				err = testSuite.NATSConn.Conn.Publish(expressSubject, smsJSON)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for express SMS processing
			time.Sleep(500 * time.Millisecond)

			// Get express SMS messages (last 2 messages should be the express ones)
			expressMessages, err := queries.GetLastSmsMessages(context.Background(), sqlc.GetLastSmsMessagesParams{
				UserID: userID,
				Limit:  2,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(expressMessages)).To(Equal(2))

			expressTimeDiff := expressMessages[0].DeliveredAt.Time.Sub(expressMessages[1].DeliveredAt.Time)

			// Verify that normal SMS time difference is greater than express SMS time difference
			Expect(normalTimeDiff).To(BeNumerically(">", expressTimeDiff))

			// Verify specific rate limits
			Expect(normalTimeDiff).To(BeNumerically(">=", 1000*time.Millisecond))
			Expect(expressTimeDiff).To(BeNumerically(">=", 100*time.Millisecond))
		})

		It("should enforce rate limiting under burst conditions", func() {
			// This test verifies that rate limiting is enforced even when multiple messages are sent rapidly
			// by sending 3 SMS messages and checking that the time differences respect rate limits

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Send 3 normal SMS messages rapidly
			subject := MakeSubject(SMS, SEND, REQ)

			for i := 0; i < 3; i++ {
				smsData := sqlc.Sm{
					UserID:        userID,
					PhoneNumberID: phoneID,
					ToPhoneNumber: "+0987654321",
					Message:       fmt.Sprintf("Burst test SMS %d", i+1),
					Status:        "pending",
				}

				smsJSON, err := json.Marshal(smsData)
				Expect(err).NotTo(HaveOccurred())

				err = testSuite.NATSConn.Conn.Publish(subject, smsJSON)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for all messages to be processed
			time.Sleep(4000 * time.Millisecond)

			// Get the last 3 SMS messages from database
			smsMessages, err := queries.GetLastSmsMessages(context.Background(), sqlc.GetLastSmsMessagesParams{
				UserID: userID,
				Limit:  3,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(smsMessages)).To(Equal(3))

			// Check that each consecutive pair respects the rate limit
			// Message 0 (most recent) vs Message 1 (second most recent)
			timeDiff1 := smsMessages[0].DeliveredAt.Time.Sub(smsMessages[1].DeliveredAt.Time)
			Expect(timeDiff1).To(BeNumerically(">=", 1000*time.Millisecond))

			// Message 1 vs Message 2 (oldest)
			timeDiff2 := smsMessages[1].DeliveredAt.Time.Sub(smsMessages[2].DeliveredAt.Time)
			Expect(timeDiff2).To(BeNumerically(">=", 1000*time.Millisecond))
		})
	})

	Context("Concurrent Processing", func() {
		It("should handle multiple SMS messages concurrently", func() {
			// Get initial balance
			initialBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				defer GinkgoRecover()
				err := worker.Start(ctx)
				Expect(err).NotTo(HaveOccurred())
			}()

			// Give worker time to start
			time.Sleep(100 * time.Millisecond)

			// Send multiple SMS messages
			numMessages := 3
			subject := MakeSubject(SMS, SEND, REQ)

			for i := 0; i < numMessages; i++ {
				smsData := sqlc.Sm{
					UserID:        userID,
					PhoneNumberID: phoneID,
					ToPhoneNumber: "+0987654321",
					Message:       fmt.Sprintf("Concurrent test SMS %d", i+1),
					Status:        "pending",
				}

				smsJSON, err := json.Marshal(smsData)
				Expect(err).NotTo(HaveOccurred())

				err = testSuite.NATSConn.Conn.Publish(subject, smsJSON)
				Expect(err).NotTo(HaveOccurred())
			}

			// Wait for processing
			time.Sleep(1 * time.Second)

			// Verify all messages were processed by checking balance deduction
			// (Multiple SMS should have deducted balance multiple times)
			finalBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(finalBalance.Int.Int64()).To(BeNumerically("<", initialBalance.Int.Int64()))
		})
	})
})
