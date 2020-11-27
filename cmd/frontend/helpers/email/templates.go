package email

type emailTemplate interface {
	filename() string
}

type ContactTemplate struct {
	IP      string
	Message string
}

func (t ContactTemplate) filename() string {
	return "contact"
}

type Forgot1Template struct {
	IP     string
	Domain string
	Code   string
}

func (t Forgot1Template) filename() string {
	return "forgot1"
}

type Forgot2Template struct {
	IP       string
	Password string
}

func (t Forgot2Template) filename() string {
	return "forgot2"
}

type ForgotMissingTemplate struct {
	IP    string
	Email string
}

func (t ForgotMissingTemplate) filename() string {
	return "forgot_missing"
}

type SignupTemplate struct {
	IP string
}

func (t SignupTemplate) filename() string {
	return "signup"
}

type VerifyTemplate struct {
	IP     string
	Domain string
	Code   string
}

func (t VerifyTemplate) filename() string {
	return "verify"
}
