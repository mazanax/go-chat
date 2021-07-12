package mailer

import (
	"fmt"
)

func PasswordRecoveryEmail(username string, email string, link string) string {
	return fmt.Sprintf(`<p>Dear, %s!</p>

<p>Somebody requested a new password for the <a href="https://chat.mznx.ru/">chat.mznx.ru</a> account associated with %s.</p>

<p><b>No changes have been made to your account yet.</b></p>

<p>You can reset your password by clicking the link below:<br>
<a href="%s">%s</a></p>

<p>This password reset link is only valid for the <b>next 5 minutes</b>.</p>

<p>If you did not request a new password, please let us know immediately by replying to this email.</p>

<p>Yours,<br>
The chat.mznx.ru team</p>`, username, email, link, link)
}
