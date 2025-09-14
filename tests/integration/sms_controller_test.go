package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/alireza-karampour/sms/internal/controllers"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/alireza-karampour/sms/tests/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SMS Controller Integration Tests", func() {
	var (
		testSuite *helpers.TestSuite
		router    *gin.Engine
		queries   *sqlc.Queries
		userID    int32
		phoneID   int32
	)

	BeforeEach(func() {
		testSuite = helpers.SetupTestSuite()
		queries = sqlc.New(testSuite.DB)

		// Setup Gin router
		gin.SetMode(gin.TestMode)
		router = gin.New()

		// Create SMS controller
		var err error
		_, err = controllers.NewSms(router.Group("/"), testSuite.DB, testSuite.NATSConn.Conn)
		Expect(err).NotTo(HaveOccurred())

		// Create test user and phone number
		balance := pgtype.Numeric{}
		balance.Scan("100.00")
		err = queries.AddUser(context.Background(), sqlc.AddUserParams{
			Username: "smstestuser",
			Balance:  balance,
		})
		Expect(err).NotTo(HaveOccurred())

		userID, err = queries.GetUserId(context.Background(), "smstestuser")
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
		testSuite.CleanupTestData()
		testSuite.Cleanup()
	})

	Context("SMS Sending", func() {
		It("should send normal SMS successfully", func() {
			// Create HTTP request
			req := httptest.NewRequest("POST", "/sms",
				helpers.JSONBody(map[string]interface{}{
					"user_id":         userID,
					"phone_number_id": phoneID,
					"to_phone_number": "+0987654321",
					"message":         "Test SMS message",
				}))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			// Parse response
			var response map[string]interface{}
			err := helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["msg"]).To(Equal("OK"))
		})

		It("should send express SMS successfully", func() {
			// Create HTTP request with express query parameter
			req := httptest.NewRequest("POST", "/sms?express=true",
				helpers.JSONBody(map[string]interface{}{
					"user_id":         userID,
					"phone_number_id": phoneID,
					"to_phone_number": "+0987654321",
					"message":         "Express SMS message",
				}))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			// Parse response
			var response map[string]interface{}
			err := helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["msg"]).To(Equal("OK"))
		})

		It("should fail to send SMS with insufficient balance", func() {
			// Create user with low balance
			lowBalance := pgtype.Numeric{}
			lowBalance.Scan("1.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: "lowbalanceuser",
				Balance:  lowBalance,
			})
			Expect(err).NotTo(HaveOccurred())

			lowBalanceUserID, err := queries.GetUserId(context.Background(), "lowbalanceuser")
			Expect(err).NotTo(HaveOccurred())

			err = queries.AddPhoneNumber(context.Background(), sqlc.AddPhoneNumberParams{
				UserID:      lowBalanceUserID,
				PhoneNumber: "+1111111111",
			})
			Expect(err).NotTo(HaveOccurred())

			lowBalancePhoneID, err := queries.GetPhoneNumberId(context.Background(), sqlc.GetPhoneNumberIdParams{
				UserID:      lowBalanceUserID,
				PhoneNumber: "+1111111111",
			})
			Expect(err).NotTo(HaveOccurred())

			balance, err := queries.GetBalance(context.Background(), lowBalanceUserID)
			Expect(err).NotTo(HaveOccurred())
			val, err := balance.MarshalJSON()
			Expect(err).NotTo(HaveOccurred())
			AddReportEntry("UserBalance: ", string(val))

			// Create HTTP request
			req := httptest.NewRequest("POST", "/sms",
				helpers.JSONBody(map[string]interface{}{
					"user_id":         lowBalanceUserID,
					"phone_number_id": lowBalancePhoneID,
					"to_phone_number": "+0987654321",
					"message":         "Test SMS message",
				}))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response - should fail with insufficient balance
			Expect(w.Code).To(Equal(http.StatusForbidden))
		})

		It("should fail with invalid JSON", func() {
			// Create HTTP request with invalid JSON
			req := httptest.NewRequest("POST", "/sms",
				helpers.JSONBody("invalid json"))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response - should fail with bad request
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("should fail with missing required fields", func() {
			// Create HTTP request with missing fields
			req := httptest.NewRequest("POST", "/sms",
				helpers.JSONBody(map[string]interface{}{
					"user_id": userID,
					// Missing phone_number_id, to_phone_number, message
				}))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response - should fail with bad request
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Balance Validation", func() {
		It("should check user balance before sending SMS", func() {
			// Get initial balance
			initialBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())

			// Send SMS
			req := httptest.NewRequest("POST", "/sms",
				helpers.JSONBody(map[string]interface{}{
					"user_id":         userID,
					"phone_number_id": phoneID,
					"to_phone_number": "+0987654321",
					"message":         "Test SMS message",
				}))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should succeed
			Expect(w.Code).To(Equal(http.StatusOK))

			// Balance should be checked (though not deducted in this implementation)
			// The actual balance deduction would happen in the worker
			currentBalance, err := queries.GetBalance(context.Background(), userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(currentBalance.Int.Int64()).To(Equal(initialBalance.Int.Int64()))
		})
	})

	Context("SMS Retrieval", func() {
		BeforeEach(func() {
			// Add some test SMS messages to the database
			err := queries.AddSms(context.Background(), sqlc.AddSmsParams{
				UserID:        userID,
				PhoneNumberID: phoneID,
				ToPhoneNumber: "+1111111111",
				Message:       "First test message",
				Status:        "delivered",
			})
			Expect(err).NotTo(HaveOccurred())

			err = queries.AddSms(context.Background(), sqlc.AddSmsParams{
				UserID:        userID,
				PhoneNumberID: phoneID,
				ToPhoneNumber: "+2222222222",
				Message:       "Second test message",
				Status:        "pending",
			})
			Expect(err).NotTo(HaveOccurred())

			err = queries.AddSms(context.Background(), sqlc.AddSmsParams{
				UserID:        userID,
				PhoneNumberID: phoneID,
				ToPhoneNumber: "+3333333333",
				Message:       "Third test message",
				Status:        "delivered",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should retrieve SMS messages for a user", func() {
			// Create HTTP request
			req := httptest.NewRequest("GET", "/sms?user_id="+helpers.Int32ToString(userID), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			// Parse response
			var response map[string]interface{}
			err := helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())

			// Check response structure
			Expect(response).To(HaveKey("messages"))
			Expect(response).To(HaveKey("count"))

			messages := response["messages"].([]interface{})
			count := response["count"].(float64)

			Expect(len(messages)).To(Equal(3))
			Expect(count).To(Equal(float64(3)))

			// Check that messages are ordered by delivered_at DESC (newest first)
			firstMessage := messages[0].(map[string]interface{})
			Expect(firstMessage["message"]).To(Equal("Third test message"))
			Expect(firstMessage["to_phone_number"]).To(Equal("+3333333333"))
		})

		It("should respect limit parameter", func() {
			// Create HTTP request with limit
			req := httptest.NewRequest("GET", "/sms?user_id="+helpers.Int32ToString(userID)+"&limit=2", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			// Parse response
			var response map[string]interface{}
			err := helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())

			messages := response["messages"].([]interface{})
			count := response["count"].(float64)

			Expect(len(messages)).To(Equal(2))
			Expect(count).To(Equal(float64(2)))
		})

		It("should use default limit when not provided", func() {
			// Create HTTP request without limit
			req := httptest.NewRequest("GET", "/sms?user_id="+helpers.Int32ToString(userID), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			// Parse response
			var response map[string]interface{}
			err := helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())

			messages := response["messages"].([]interface{})
			count := response["count"].(float64)

			// Should return all 3 messages (default limit is 10, but we only have 3)
			Expect(len(messages)).To(Equal(3))
			Expect(count).To(Equal(float64(3)))
		})

		It("should enforce maximum limit", func() {
			// Create HTTP request with limit exceeding maximum
			req := httptest.NewRequest("GET", "/sms?user_id="+helpers.Int32ToString(userID)+"&limit=200", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			// Parse response
			var response map[string]interface{}
			err := helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())

			messages := response["messages"].([]interface{})
			count := response["count"].(float64)

			// Should be limited to 100 (max limit)
			Expect(len(messages)).To(Equal(3)) // We only have 3 messages
			Expect(count).To(Equal(float64(3)))
		})

		It("should fail with missing user_id", func() {
			// Create HTTP request without user_id
			req := httptest.NewRequest("GET", "/sms", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response - should fail with bad request
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("should return empty list for user with no messages", func() {
			// Create another user with no messages
			balance := pgtype.Numeric{}
			balance.Scan("50.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: "emptyuser",
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())

			emptyUserID, err := queries.GetUserId(context.Background(), "emptyuser")
			Expect(err).NotTo(HaveOccurred())

			// Create HTTP request for user with no messages
			req := httptest.NewRequest("GET", "/sms?user_id="+helpers.Int32ToString(emptyUserID), nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))

			// Parse response
			var response map[string]interface{}
			err = helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())

			messages := response["messages"].([]interface{})
			count := response["count"].(float64)

			Expect(len(messages)).To(Equal(0))
			Expect(count).To(Equal(float64(0)))
		})
	})
})
