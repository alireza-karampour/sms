package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/alireza-karampour/sms/internal/subjects"
	. "github.com/alireza-karampour/sms/pkg/utils"
)

var _ = Describe("Utils", func() {
	Context("HasSubject", func() {
		It("should handle * correctly", func() {
			msgSubject := Subject("sms.send.request")
			res := msgSubject.HasSubject(SMS, "*", REQ)
			Expect(res).To(BeTrue())
		})
		It("should fail", func() {
			msgSubject := Subject("sms.send.request")
			res := msgSubject.HasSubject(SMS, EX, SEND, REQ)
			Expect(res).To(BeFalse())
		})
		It("should fail", func() {
			msgSubject := Subject("sms.send.request")
			res := msgSubject.HasSubject(SMS, EX, REQ)
			Expect(res).To(BeFalse())
		})
	})
})
