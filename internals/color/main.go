package color

const Reset string = "\033[0m"
const Red string = "\033[31m"
const Green string = "\033[32m"
const Yellow string = "\033[33m"
const Blue string = "\033[34m"
const Magenta string = "\033[35m"
const Cyan string = "\033[36m"
const Gray string = "\033[37m"
const White string = "\033[97m"

func RedString(str string) string {
	return Red + str + Reset
}
func GreenString(str string) string {
	return Green + str + Reset
}
func YellowString(str string) string {
	return Yellow + str + Reset
}
func BlueString(str string) string {
	return Blue + str + Reset
}
func MagentaString(str string) string {
	return Magenta + str + Reset
}
func CyanString(str string) string {
	return Cyan + str + Reset
}
func GrayString(str string) string {
	return Gray + str + Reset
}
func WhiteString(str string) string {
	return White + str + Reset
}
