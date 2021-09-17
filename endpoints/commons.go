package endpoints

var NoConsole bool = true

// Return Error ? true : false
func CheckError(err error) bool {
	if err != nil {
		Print("Error : " + err.Error())
		return true
	}
	return false
}

func PrintError(err error) {
	if err != nil {
		Print(err.Error())
	}
}

func PrintErrorMsg(msg string, err error) {
	if err != nil {
		Print(msg + "\t" + err.Error())
	}
}

func Print(s string) {
	println(s)
}
