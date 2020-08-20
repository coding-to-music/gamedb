package session

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gorilla/securecookie"
	"go.uber.org/zap"
)

func Get(r *http.Request, key string) (value string) {

	val, err := session.Get(r, key)
	logSessionError(err)
	return val
}

func Set(r *http.Request, name string, value string) {

	err := session.Set(r, name, value)
	logSessionError(err)
}

func SetMany(r *http.Request, values map[string]string) {

	err := session.SetMany(r, values)
	logSessionError(err)
}

func GetFlashes(r *http.Request, group FlashGroup) (flashes []string) {

	flashes, err := session.GetFlashes(r, session.FlashGroup(group))
	logSessionError(err)

	return flashes
}

func SetFlash(r *http.Request, group FlashGroup, flash string) {

	err := session.SetFlash(r, session.FlashGroup(group), flash)
	logSessionError(err)
}

func DeleteMany(r *http.Request, keys []string) {

	err := session.DeleteMany(r, keys)
	logSessionError(err)
}

func DeleteAll(r *http.Request) {

	err := session.DeleteAll(r)
	logSessionError(err)
}

func Save(w http.ResponseWriter, r *http.Request) {

	err := session.Save(w, r)
	logSessionError(err)
}

func logSessionError(err error) {

	if err != nil {

		if val, ok := err.(securecookie.Error); ok {
			if val.IsUsage() || val.IsDecode() {
				zap.S().Info(val.Error())
				return
			}
		}

		zap.S().Error(err)
	}
}
