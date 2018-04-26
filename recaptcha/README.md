# Recaptcha

```go
package main

import "github.com/Jleagle/recaptcha-go"

err := recaptcha.CheckFromRequest(r)
if err != nil {
    e, ok := err.(recaptcha.Error)
    if ok {
        if e.IsUserError() {
            return err.Error()
        }else{
            logger.Error(e)
        }
    }
}
```
