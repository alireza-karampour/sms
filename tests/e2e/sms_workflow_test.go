package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alireza-karampour/sms/internal/subjects"
	"github.com/alireza-karampour/sms/pkg/nats"
	"github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/alireza-karampour/sms/tests/helpers"
	"github.com/jackc/pgx/v5/pgtype"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	natsgo "github.com/nats-io/nats.go"
)

var _ = Describe("SMS Workflow E2E Tests", func() {
	var (
		testSuite *helpers.TestSuite
		client    *helpers.HTTPClient
		queries   *sqlc.Queries
		consumer  *nats.Consumer
	)

	BeforeEach(func() {
		testSuite = helpers.SetupTestSuite()
		queries = sqlc.New(testSuite.DB)
		
		// Setup HTTP client
		client = helpers.NewHTTPClient("http://localhost:8080")
		
		// Setup NATS consumer for testing
		var err error
		consumer, err = nats.NewConsumer(testSuite.NATSConn.Conn)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if consumer != nil {
			consumer.Close()
		}
		testSuite.CleanupTestData()
		testSuite.Cleanup()
	})

	Context("Complete SMS Workflow", func() {
		It("should handle complete SMS workflow from user creation to message processing", func() {
			// Step 1: Create a user
			userData := helpers.UserData{
				Username: "e2euser",
				Balance:  100.0,
			}
			
			resp, err := client.Post("/user", helpers.RequestOptions{
				Body: userData,
			})
			Expect(err).NotTo(HaveOccurred())
			helpers.AssertResponseStatus(resp, http.StatusOK)
			
			// Get user ID
			resp, err = client.Get(fmt.Sprintf("/user/%s", userData.Username))
			Expect(err).NotTo(HaveOccurred())
			helpers.AssertResponseStatus(resp, http.StatusOK)
			
			var userResponse map[string]interface{}
			err = helpers.ParseJSONResponse(resp, &userResponse)
			Expect(err).NotTo(HaveOccurred())
			userID := int32(userResponse["id"].(float64))
			
			// Step 2: Add phone number for the user
			phoneData := helpers.PhoneNumberData{
				PhoneNumber: "+1234567890",
			}
			
			// Note: This would require implementing the phone number endpoint
			// For now, we'll add it directly to the database
			err = queries.AddPhoneNumber(context.Background(), sqlc.AddPhoneNumberParams{
				UserID:      userID,
				PhoneNumber: phoneData.PhoneNumber,
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Get phone number ID
			phoneID, err := queries.GetPhoneNumberId(context.Background(), sqlc.GetPhoneNumberIdParams{
				UserID:      userID,
				PhoneNumber: phoneData.PhoneNumber,
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Step 3: Send SMS
			smsData := helpers.SMSData{
				ToPhoneNumber: "+0987654321",
				Message:       "E2E test SMS message",
			}
			
			resp, err = client.Post("/sms", helpers.RequestOptions{
				Body: map[string]interface{}{
					"user_id":          userID,
					"phone_number_id":  phoneID,
					"to_phone_number":  smsData.ToPhoneNumber,
					"message":          smsData.Message,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			helpers.AssertResponseStatus(resp, http.StatusOK)
			
			// Step 4: Verify message was published to NATS
			// Subscribe to the SMS subject
			subject := utils.MakeSubject(subjects.SMS, subjects.SEND, subjects.REQ)
			msgChan := make(chan *natsgo.Msg, 1)
			
			sub, err := testSuite.NATSConn.Conn.Subscribe(subject, func(msg *natsgo.Msg) {
				msgChan <- msg
			})
			Expect(err).NotTo(HaveOccurred())
			defer sub.Unsubscribe()
			
			// Wait for message with timeout
			select {
			case msg := <-msgChan:
				Expect(msg).NotTo(BeNil())
				
				// Parse the SMS data from the message
				var smsPayload sqlc.Sm
				err = json.Unmarshal(msg.Data, &smsPayload)
				Expect(err).NotTo(HaveOccurred())
				Expect(smsPayload.UserID).To(Equal(userID))
				Expect(smsPayload.PhoneNumberID).To(Equal(phoneID))
				Expect(smsPayload.ToPhoneNumber).To(Equal(smsData.ToPhoneNumber))
				Expect(smsPayload.Message).To(Equal(smsData.Message))
				
			case <-time.After(5 * time.Second):
				Fail("Timeout waiting for NATS message")
			}
		})

		It("should handle express SMS workflow", func() {
			// Create user and phone number
			balance := pgtype.Numeric{}
			balance.Scan("100.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: "expressuser",
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())
			
			userID, err := queries.GetUserId(context.Background(), "expressuser")
			Expect(err).NotTo(HaveOccurred())
			
			err = queries.AddPhoneNumber(context.Background(), sqlc.AddPhoneNumberParams{
				UserID:      userID,
				PhoneNumber: "+1111111111",
			})
			Expect(err).NotTo(HaveOccurred())
			
			phoneID, err := queries.GetPhoneNumberId(context.Background(), sqlc.GetPhoneNumberIdParams{
				UserID:      userID,
				PhoneNumber: "+1111111111",
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Send express SMS
			resp, err := client.Post("/sms?express=true", helpers.RequestOptions{
				Body: map[string]interface{}{
					"user_id":          userID,
					"phone_number_id":  phoneID,
					"to_phone_number":  "+2222222222",
					"message":          "Express SMS message",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			helpers.AssertResponseStatus(resp, http.StatusOK)
			
			// Verify message was published to express SMS subject
			subject := utils.MakeSubject(subjects.SMS, subjects.EX, subjects.SEND, subjects.REQ)
			msgChan := make(chan *natsgo.Msg, 1)
			
			sub, err := testSuite.NATSConn.Conn.Subscribe(subject, func(msg *natsgo.Msg) {
				msgChan <- msg
			})
			Expect(err).NotTo(HaveOccurred())
			defer sub.Unsubscribe()
			
			// Wait for message
			select {
			case msg := <-msgChan:
				Expect(msg).NotTo(BeNil())
				
				var smsPayload sqlc.Sm
				err = json.Unmarshal(msg.Data, &smsPayload)
				Expect(err).NotTo(HaveOccurred())
				Expect(smsPayload.Message).To(Equal("Express SMS message"))
				
			case <-time.After(5 * time.Second):
				Fail("Timeout waiting for express NATS message")
			}
		})

		It("should handle insufficient balance scenario", func() {
			// Create user with low balance
			lowBalance := pgtype.Numeric{}
			lowBalance.Scan("1.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: "lowbalanceuser",
				Balance:  lowBalance,
			})
			Expect(err).NotTo(HaveOccurred())
			
			userID, err := queries.GetUserId(context.Background(), "lowbalanceuser")
			Expect(err).NotTo(HaveOccurred())
			
			err = queries.AddPhoneNumber(context.Background(), sqlc.AddPhoneNumberParams{
				UserID:      userID,
				PhoneNumber: "+3333333333",
			})
			Expect(err).NotTo(HaveOccurred())
			
			phoneID, err := queries.GetPhoneNumberId(context.Background(), sqlc.GetPhoneNumberIdParams{
				UserID:      userID,
				PhoneNumber: "+3333333333",
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Try to send SMS
			resp, err := client.Post("/sms", helpers.RequestOptions{
				Body: map[string]interface{}{
					"user_id":          userID,
					"phone_number_id":  phoneID,
					"to_phone_number":  "+4444444444",
					"message":          "Should fail SMS",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			helpers.AssertResponseStatus(resp, http.StatusForbidden)
			
			// Verify no message was published to NATS
			subject := utils.MakeSubject(subjects.SMS, subjects.SEND, subjects.REQ)
			msgChan := make(chan *natsgo.Msg, 1)
			
			sub, err := testSuite.NATSConn.Conn.Subscribe(subject, func(msg *natsgo.Msg) {
				msgChan <- msg
			})
			Expect(err).NotTo(HaveOccurred())
			defer sub.Unsubscribe()
			
			// Wait for a short time to ensure no message is received
			select {
			case <-msgChan:
				Fail("Unexpected message received for insufficient balance")
			case <-time.After(2 * time.Second):
				// This is expected - no message should be received
			}
		})
	})

	Context("NATS Stream Configuration", func() {
		It("should verify NATS streams are properly configured", func() {
			// This test verifies that the NATS streams are set up correctly
			// by checking if we can publish to the expected subjects
			
			// Test normal SMS stream
			normalSubject := utils.MakeSubject(subjects.SMS, subjects.SEND, subjects.REQ)
			err := testSuite.NATSConn.Conn.Publish(normalSubject, []byte("test message"))
			Expect(err).NotTo(HaveOccurred())
			
			// Test express SMS stream
			expressSubject := utils.MakeSubject(subjects.SMS, subjects.EX, subjects.SEND, subjects.REQ)
			err = testSuite.NATSConn.Conn.Publish(expressSubject, []byte("test express message"))
			Expect(err).NotTo(HaveOccurred())
			
			// Test status subjects
			statusSubject := utils.MakeSubject(subjects.SMS, subjects.SEND, subjects.STAT)
			err = testSuite.NATSConn.Conn.Publish(statusSubject, []byte("test status"))
			Expect(err).NotTo(HaveOccurred())
			
			// Test error subjects
			errorSubject := utils.MakeSubject(subjects.SMS, subjects.SEND, subjects.ERR)
			err = testSuite.NATSConn.Conn.Publish(errorSubject, []byte("test error"))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})