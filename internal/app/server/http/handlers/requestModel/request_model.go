package requestModel

import (
	"fmt"
	"regexp"
)

type ReqLink struct {
	LinkName  string `json:"link_name"`
	LinkColor string `json:"link_color"`
	LinkPath  string `json:"link_path"`
}

type ReqUpdateLink struct {
	LinkID    int    `json:"link_id"`
	LinkName  string `json:"link_name"`
	LinkColor string `json:"link_color"`
	LinkPath  string `json:"link_path"`
}

type SignUpModel struct {
	Email    string `json:"email"`
	Username string `json:"login"`
	Password string `json:"password"`
}

type LoginModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (sm *SignUpModel) Validate() error {
	if ok := isValidEmail(sm.Email); !ok {
		return fmt.Errorf("invalid email")
	}

	if ok := isValidUsername(sm.Username); !ok {
		return fmt.Errorf("invalid login")
	}

	return nil
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
}

func isValidUsername(username string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,20}$`, username)
	return matched
}
