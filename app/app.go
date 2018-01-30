package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/qinhao/ppsh"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

var (
	Name      = "ppsh"
	Usage     = "Pull or Push via SSH in your cluster hosts."
	Version   = "0.0.1"
	Author    = "qinhao"
	Email     = "qinhao@qinhao.me"
	BuildTime = "2000-02-02T00:00:00+0800"
)

var (
	logo = `
██████╗ ██████╗ ███████╗██╗  ██╗
██╔══██╗██╔══██╗██╔════╝██║  ██║
██████╔╝██████╔╝███████╗███████║
██╔═══╝ ██╔═══╝ ╚════██║██╔══██║
██║     ██║     ███████║██║  ██║
╚═╝     ╚═╝     ╚══════╝╚═╝  ╚═╝`
)

func main() {
	app := cli.NewApp()

	flags := []cli.Flag{
		stringFlag("hosts, H", "", "host list, in the form of `HOST[;HOST]`"),
		stringFlag("cmds, c", "", "command list, in the form of `CMD[;CMD]`"),
		stringFlag("ciphers, C", "", "cipher list, in the form of `CIPHER[;CIPHER]`"),
		stringFlag("ip-range, I", "", "ip range, in the form of `IP-IP[;{IP-IP|IP/XX}`"),
		stringFlag("user, u", "root", "ssh login `USER`"),
		stringFlag("password, w", "", "ssh login `PASSWORD`"),
		stringFlag("cert-key, k", "", "ssh private key `FILE`"),
		stringFlag("playbook, p", "", "load playbook from path `FILE`"),
		stringFlag("taskbook, t", "", "load taskbook from path `FILE`"),
		stringFlag("format, f", "plain", "output as `PLAIN|JSON`"),
		stringFlag("platform, S", "linux", "platform as `LINUX|OTHER`"),
		stringFlag("output, o", "stdout", "output to `STDOUT|FILE`"),
		intFlag("timeout, s", 30, "`TIMEOUT` in second"),
		intFlag("port, P", 22, "ssh `PORT`"),
		intFlag("max-run-count, n", 20, "max runing `COUNT`"),
	}

	app.Name = Name
	app.Usage = Usage
	app.Flags = append(app.Flags, flags...)
	app.Version = Version
	app.Action = action
	app.EnableBashCompletion = true

	app.Run(os.Args)
}

func action(c *cli.Context) error {
	if c.NumFlags() == 0 {
		color.Blue("run `ppsh help` for more usage ;-)")
		os.Exit(0)
	}

	hosts := splitArg(c.String("hosts"))
	cmds := splitArg(c.String("cmds"))
	ciphers := splitArg(c.String("ciphers"))

	if c.String("ip-range") != "" {
		hosts = ppsh.ParseIPRange(c.String("ip-range"))
	}

	var platformO = ppsh.LINUX
	if c.String("platform") == "other" {
		platformO = ppsh.OTHER
	}

	if c.String("taskbook") != "" {
		var err error
		cmds, err = parseTaskbook(c.String("taskbook"))
		if err != nil {
			fmt.Printf("parse taskbook error[%v]", err)
			os.Exit(1)
		}
	}

	var playbookO = ppsh.Playbook{
		Format: ppsh.PLAIN, // set default format plain
		MaxNum: c.Int("max-run-count"),
		Out:    os.Stdout, //todo: more out types
	}

	if c.String("format") == "json" {
		playbookO.Format = ppsh.JSON
	}

	if c.String("playbook") != "" {
		playbookO.Path = c.String("playbook")
		err := playbookO.Parse()
		if err != nil {
			fmt.Printf("open playbook err[%v]", err)
			os.Exit(1)
		}
	} else {
		shosts := []ppsh.Host{}

		for _, host := range hosts {
			h := ppsh.Host{
				IP:       host,
				Port:     c.Int("port"),
				User:     c.String("user"),
				Password: c.String("password"),
				CertKey:  c.String("cert-key"),
				Ciphers:  ciphers,
				Tasks:    cmds,
				Platform: platformO,
				Timeout:  c.Int("timeout"),
			}

			shosts = append(shosts, h)
		}
		playbookO.Hosts = shosts
	}

	//fmt.Printf("%v", playbookO.Hosts)

	layout := "2006-01-02T15:04:05.999999-07:00"
	startTime := time.Now()
	color.Blue("ppsh start:\t\t%s", color.GreenString(startTime.Format(layout)))

	results := playbookO.Play()

	endTime := time.Now()
	color.Blue("ppsh finished:\t\t%s", color.GreenString(endTime.Format(layout)))

	output(results, playbookO.Format, endTime.Sub(
		startTime), len(playbookO.Hosts))

	return nil
}

func stringFlag(name, value, usage string) cli.StringFlag {
	return cli.StringFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	}
}

func boolFlag(name, usage string) cli.BoolFlag {
	return cli.BoolFlag{
		Name:  name,
		Usage: usage,
	}
}

func intFlag(name string, value int, usage string) cli.IntFlag {
	return cli.IntFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	}
}

func durationFlag(name string, value time.Duration, usage string) cli.DurationFlag {
	return cli.DurationFlag{
		Name:  name,
		Value: value,
		Usage: usage,
	}
}

func splitArg(s string) (ss []string) {
	if s == "" {
		return
	}
	if strings.Contains(s, ";") {
		return strings.Split(s, ";")
	}

	ss = append(ss, s)
	return
}

func parseTaskbook(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var tasks []string
	if err := yaml.Unmarshal([]byte(data), &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func output(rs []ppsh.Result, format ppsh.Format, spent time.Duration, number int) {
	color.Blue("proccess time:\t\t%s", color.GreenString("%s", spent))
	color.Blue("number of hosts:\t%s", color.GreenString("%d", number))

	color.Red("OUTPUT")

	var outputPlain = func(r *ppsh.Result) {
		breakln()
		color.Blue("host:\t\t\t%s", color.GreenString(r.Host))
		color.Blue("cmd:\t\t\t%s", color.GreenString(r.Cmd))
		if r.Code != 0 {
			color.Blue("code:\t\t\t%s", color.GreenString("%d", r.Code))
		}
		if r.Detail != "" {
			color.Blue("detail:\t\t\t%s", color.GreenString("%q", r.Detail))
		}
		if r.Error != "" {
			color.Blue("error:\t\t\t%s", color.GreenString(r.Error))
		}
	}

	if format == ppsh.PLAIN {
		for _, r := range rs {
			outputPlain(&r)
		}
		breakln()
		return
	}

	for _, r := range rs {
		result, err := json.Marshal(r)
		if err != nil {
			color.Red("parse rusult to json error: %v, just output in plain:", err)
			outputPlain(&r)
			break
		}

		var out bytes.Buffer
		err = json.Indent(&out, result, "", "\t")

		breakln()
		color.Blue("host:\t\t\t%s", color.GreenString(r.Host))
		color.Blue("result:\n%s", color.GreenString(out.String()))
	}
	breakln()
}

func breakln() {
	color.Blue("--------------------------------------------------------")
}
