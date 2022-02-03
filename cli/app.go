package cli

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/thomaso-mirodin/intmath/intgr"
)

type cmd struct {
	name   string
	abbrev string
	fn     func() error
}

type app struct {
	cmds    []*cmd
	actions []string
}

func makeApp() *app {
	return &app{}
}

func (a *app) Register(name string, fn func() error) {
	c := &cmd{
		name: name,
		fn:   fn,
	}
	a.cmds = append(a.cmds, c)
}

func (a *app) Init() error {
	var actionList []string

	// Pull arguments before flags into the acton map
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		lastCommand := 1
		for lastCommand < len(os.Args) {
			if action := strings.TrimSpace(strings.ToLower(os.Args[lastCommand])); action != "" && !strings.HasPrefix(action, "-") {
				actionList = append(actionList, action)
				lastCommand++
			} else {
				break
			}
		}
		var newArgs []string
		for i, arg := range os.Args {
			if i == 0 || i >= lastCommand {
				newArgs = append(newArgs, arg)
			}
		}
		os.Args = newArgs
	}

	// Parse flags after modifying os.Args
	flag.Parse()

	if *actions != "" {
		for _, c := range strings.Split(*actions, ",") {
			if action := strings.TrimSpace(strings.ToLower(c)); action != "" {
				actionList = append(actionList, action)
			}
		}
	}
	for _, c := range flag.Args() {
		if action := strings.TrimSpace(strings.ToLower(c)); action != "" {
			actionList = append(actionList, action)
		}
	}

	if len(actionList) == 0 {
		return errors.Errorf("you need to specify at least one call")
	}

	a.actions = actionList

	return nil
}

func (a *app) findCmd(s string) *cmd {
	for _, c := range a.cmds {
		if strings.EqualFold(s, c.name) {
			return c
		}
		if strings.EqualFold(s, c.abbrev) {
			return c
		}
	}
	return nil
}

func (a *app) ShowHelp() {
	repeat := func(n int) string {
		var res string
		for i := 0; i < n; i++ {
			res += "="
		}
		return res
	}
	var namePad int
	{
		maxNameLength := math.MinInt
		for _, c := range a.cmds {
			maxNameLength = intgr.Max(maxNameLength, len(c.name))
		}
		namePad = maxNameLength + 2
	}

	fmt.Println("The following commands are available:")
	fmt.Println()
	fmt.Printf("  %"+fmt.Sprintf("%d", namePad)+"s - %s\n", "Action", "Abbreviation")
	fmt.Printf("  %"+fmt.Sprintf("%d", namePad)+"s - %s\n", repeat(namePad), repeat(len("Abbreviation")))
	for _, c := range a.cmds {
		fmt.Printf("  %"+fmt.Sprintf("%d", namePad)+"s - %s\n", c.name, c.abbrev)
	}
}

func (a *app) preRun() {
	sort.Slice(a.cmds, func(i, j int) bool {
		return a.cmds[i].name < a.cmds[j].name
	})
	getAbbrev := func(name string) string {
		var buf bytes.Buffer
		for _, s := range name {
			s := string(s)
			if strings.ToUpper(s) == s {
				buf.WriteString(strings.ToLower(s))
			}
		}
		return buf.String()
	}
	isUnique := func(s string) bool {
		for _, c := range a.cmds {
			if c.abbrev == s {
				return false
			}
		}
		return true
	}
	for _, c := range a.cmds {
		abbrev := getAbbrev(c.name)
		if !isUnique(abbrev) {
			for i := 0; i < len(c.name); i++ {
				sub := strings.ToLower(string(c.name[0:i]))
				if isUnique(sub) {
					abbrev = sub
					break
				}
			}
		}
		if !isUnique(abbrev) {
			abbrev = strings.ToLower(c.name)
		}
		c.abbrev = abbrev
	}
}

func (a *app) Run() error {
	a.preRun()
	for _, s := range a.actions {
		c := a.findCmd(s)
		if c == nil {
			return errors.Errorf("no action for %q", s)
		}
		if err := c.fn(); err != nil {
			return errors.Errorf("running %q: %v", c.name, err)
		}
	}
	return nil
}
