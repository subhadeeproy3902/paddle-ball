package game

// title_art.go - compact single-line figlet wordmark (font: small) for the
// title screen. Kept short (4 rows, fits one line on any >=60-col terminal)
// and gradient-colored at render time in view.go.

var titleWordmark = []string{
	" ___  _   ___  ___  _    ___   ___   _   _    _    ",
	"| _ \\/_\\ |   \\|   \\| |  | __| | _ ) /_\\ | |  | |   ",
	"|  _/ _ \\| |) | |) | |__| _|  | _ \\/ _ \\| |__| |__ ",
	"|_|/_/ \\_\\___/|___/|____|___| |___/_/ \\_\\____|____|",
}
