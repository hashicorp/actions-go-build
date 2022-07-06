package cli

type Args interface {
	ParseArgs([]string) error
}
