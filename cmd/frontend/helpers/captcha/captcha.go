package captcha

import (
	"sync"

	"github.com/Jleagle/captcha-go"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	hcaptchaClient captcha.Provider
	hcaptchaLock   sync.Mutex
)

func Client() captcha.Provider {

	hcaptchaLock.Lock()
	defer hcaptchaLock.Unlock()

	if hcaptchaClient == nil {
		hcaptchaClient = captcha.New(captcha.HCaptcha, config.C.HCaptchaSecret, config.C.HCaptchaPublic)
	}

	return hcaptchaClient
}
