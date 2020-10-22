package email

type Template string

const (
	TemplateContact = "contact"
	TemplateForgot1 = "forgot1"
	TemplateForgot2 = "forgot2"
	TemplateSignup  = "signup"
	TemplateVerify  = "verify"
)

type ContactTemplate struct {
	Message string
}

type Forgot1Template struct {
	Domain string
	Code   string
	IP     string
}

type Forgot2Template struct {
	Password string
	IP       string
}

type SignupTemplate struct {
	IP string
}

type VerifyTemplate struct {
	Domain string
	Code   string
}
