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
	IP      string
	Message string
}

type Forgot1Template struct {
	IP     string
	Domain string
	Code   string
}

type Forgot2Template struct {
	IP       string
	Password string
}

type SignupTemplate struct {
	IP string
}

type VerifyTemplate struct {
	IP     string
	Domain string
	Code   string
}
