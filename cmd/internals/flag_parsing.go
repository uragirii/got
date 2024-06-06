package internals

type flagType int

const (
	Bool flagType = iota
	String
)

type Flag struct {
	Name  string
	Short string
	Help  string
	Key   string
	Type  flagType
}

type Command struct {
	Name       string
	Desc       string
	Flags      []*Flag
	Run        func(c *Command)
	Args       []string
	parsedFlag map[string]string
}

func (c *Command) ParseCommand(args []string) {
	c.parsedFlag = make(map[string]string)
	c.Args = make([]string, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg[0] == '-' {
			c.parseFlag(args[i:])
		} else {
			c.Args = append(c.Args, arg)
		}
	}
}

func getRawFlag(flag string) string {
	if flag[0:2] == "--" {
		return flag[2:]
	}

	return flag[1:]
}

func (c *Command) parseFlag(args []string) {
	flag := getRawFlag(args[0])
	rest := args[1:]

	for _, commandFlag := range c.Flags {
		if commandFlag.Name == flag || commandFlag.Short == flag {
			switch commandFlag.Type {
			case Bool:
				c.parsedFlag[commandFlag.Key] = "true"
			case String:
				c.parsedFlag[commandFlag.Key] = rest[0]
			}
		}
	}
}

func (c *Command) GetFlag(flag string) string {
	return c.parsedFlag[flag]
}
