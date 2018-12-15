package format

import "regexp"

var validCellphone = regexp.MustCompile(`^1\d{10}$`)

//var validEmail = regexp.MustCompile(`^[0-9a-zA-Z][a-zA-Z0-9\._-]{1,}@[a-z0-9-]{1,}[a-z0-9]\.[a-z\.]{1,}[a-z]$`)
//var validEmail = regexp.MustCompile(`^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+((\.[a-zA-Z0-9_-]{2,3}){1,2})$`)
var validEmail = regexp.MustCompile(`^\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
var validID = regexp.MustCompile(`^(\d{6})(18|19|20)(\d{2})([01]\d)([0123]\d)(\d{3})(\d|X)$`)
var validPassword = regexp.MustCompile(`^[\w\+\.\*\(\)-_]{6,16}$`)

func CheckCellphone(phoneNumber string) bool {
	return validCellphone.MatchString(phoneNumber)
}

func CheckEmail(email string) bool {
	return validEmail.MatchString(email)
}

func CheckIDCard(id string) bool {
	return validID.MatchString(id)
}

func CheckPassword(password string) bool {
	return validPassword.MatchString(password)
}
