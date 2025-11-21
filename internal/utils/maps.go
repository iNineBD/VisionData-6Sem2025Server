package utils

// UserTypMapIntToStr maps user type integers to their string representations
var UserTypMapIntToStr = map[int]string{
	1: "ADMIN",
	2: "MANAGER",
	3: "SUPPORT",
}

// UserTypMapStrToInt maps user type strings to their integer representations
var UserTypMapStrToInt = map[string]int{
	"ADMIN":   1,
	"MANAGER": 2,
	"SUPPORT": 3,
}
