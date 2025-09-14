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

var _ = Describe("User Controller Integration Tests", func() {
	var (
		testSuite *helpers.TestSuite
		router    *gin.Engine
		queries   *sqlc.Queries
	)

	BeforeEach(func() {
		testSuite = helpers.SetupTestSuite()
		queries = sqlc.New(testSuite.DB)
		
		// Setup Gin router
		gin.SetMode(gin.TestMode)
		router = gin.New()
		
		// Create user controller
		_ = controllers.NewUser(router.Group("/"), testSuite.DB)
	})

	AfterEach(func() {
		testSuite.CleanupTestData()
		testSuite.Cleanup()
	})

	Context("User Creation", func() {
		It("should create a new user successfully", func() {
			// Test data
			username := "testuser"
			balance := pgtype.Numeric{}
			balance.Scan("100.00")
			
			// Create user via database
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: username,
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Verify user was created
			userID, err := queries.GetUserId(context.Background(), username)
			Expect(err).NotTo(HaveOccurred())
			Expect(userID).To(BeNumerically(">", 0))
		})

		It("should fail to create user with duplicate username", func() {
			username := "duplicateuser"
			balance := pgtype.Numeric{}
			balance.Scan("100.00")
			
			// Create first user
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: username,
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Try to create duplicate user
			err = queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: username,
				Balance:  balance,
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("User Retrieval", func() {
		var userID int32

		BeforeEach(func() {
			balance := pgtype.Numeric{}
			balance.Scan("150.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: "testuser",
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())
			
			userID, err = queries.GetUserId(context.Background(), "testuser")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should retrieve user by username", func() {
			id, err := queries.GetUserId(context.Background(), "testuser")
			Expect(err).NotTo(HaveOccurred())
			Expect(id).To(Equal(userID))
		})

		It("should return error for non-existent user", func() {
			_, err := queries.GetUserId(context.Background(), "nonexistent")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("User Balance Management", func() {
		var username string

		BeforeEach(func() {
			username = "testuser"
			balance := pgtype.Numeric{}
			balance.Scan("100.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: username,
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should add balance correctly", func() {
			amount := pgtype.Numeric{}
			amount.Scan("50.00")
			newBalance, err := queries.AddBalance(context.Background(), sqlc.AddBalanceParams{
				Username: username,
				Balance:  amount,
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Verify balance was updated (100.00 + 50.00 = 150.00)
			expectedBalance := pgtype.Numeric{}
			expectedBalance.Scan("150.00")
			Expect(newBalance.Int.Int64).To(Equal(expectedBalance.Int.Int64))
		})
	})

	Context("HTTP API Tests", func() {
		It("should create user via HTTP POST", func() {
			// Create HTTP request
			req := httptest.NewRequest("POST", "/user", 
				helpers.JSONBody(map[string]interface{}{
					"username": "httptestuser",
					"balance":  "100.00",
				}))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Perform request
			router.ServeHTTP(w, req)
			
			// Assert response
			Expect(w.Code).To(Equal(http.StatusOK))
			
			// Verify user was created in database
			userID, err := queries.GetUserId(context.Background(), "httptestuser")
			Expect(err).NotTo(HaveOccurred())
			Expect(userID).To(BeNumerically(">", 0))
		})

		It("should get user ID via HTTP GET", func() {
			// First create a user
			balance := pgtype.Numeric{}
			balance.Scan("100.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: "gettestuser",
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Create HTTP request
			req := httptest.NewRequest("GET", "/user/gettestuser", nil)
			
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
			Expect(response["id"]).To(BeNumerically(">", 0))
		})

		It("should add balance via HTTP PUT", func() {
			// First create a user
			username := "balancetestuser"
			balance := pgtype.Numeric{}
			balance.Scan("100.00")
			err := queries.AddUser(context.Background(), sqlc.AddUserParams{
				Username: username,
				Balance:  balance,
			})
			Expect(err).NotTo(HaveOccurred())
			
			// Create HTTP request
			req := httptest.NewRequest("PUT", "/user/balance", 
				helpers.JSONBody(map[string]interface{}{
					"username": username,
					"balance":  "50.00",
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
			err = helpers.ParseJSONResponse(w.Result(), &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["status"]).To(Equal(float64(200)))
			Expect(response["new_balance"]).To(Equal("150.00"))
		})
	})
})