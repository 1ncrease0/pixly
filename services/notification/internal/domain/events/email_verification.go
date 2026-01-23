package events

type EmailVerification struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
