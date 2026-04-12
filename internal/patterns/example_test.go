package patterns_test

import (
	"errors"
	"fmt"

	"github.com/example/go-template/internal/patterns"
)

func ExampleNewServer() {
	s := patterns.NewServer("0.0.0.0",
		patterns.WithPort(9090),
		patterns.WithMaxConns(50),
	)
	fmt.Printf("addr=%s port=%d conns=%d\n", s.Addr, s.Port, s.MaxConns)
	// Output:
	// addr=0.0.0.0 port=9090 conns=50
}

func ExampleValidate() {
	err := patterns.Validate("")
	var ve *patterns.ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("field=%s msg=%s\n", ve.Field, ve.Message)
	}
	// Output:
	// field=name msg=required
}

func ExampleWrap() {
	err := patterns.Wrap(patterns.ErrNotFound, "user lookup")
	fmt.Println(err)
	fmt.Println(errors.Is(err, patterns.ErrNotFound))
	// Output:
	// user lookup: not found
	// true
}

func ExampleMultiNotifier() {
	mn := patterns.MultiNotifier{
		patterns.EmailNotifier{Addr: "<email>"},
		patterns.SlackNotifier{Webhook: "https://hooks.example.com/x"},
	}
	_ = mn.Notify("deploy complete")
	// Output:
	// [Email→<email>] deploy complete
	// [Slack→https://hooks.example.com/x] deploy complete
}
