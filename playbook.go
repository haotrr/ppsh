package ppsh

import (
	"io"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Playbook is a book to be played.
type Playbook struct {
	Path   string
	Hosts  []Host
	Out    io.Writer
	Format Format
	MaxNum int
}

// NewPlaybook return a new playbook.
func NewPlaybook(path string, out io.Writer, format Format, max int) *Playbook {
	return &Playbook{
		Path:   path,
		Out:    out,
		Format: format,
		MaxNum: max,
	}
}

// Parse parse the playbook.
func (p *Playbook) Parse() error {
	f, err := os.Open(p.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)

	err = yaml.Unmarshal([]byte(data), &p.Hosts)
	if err != nil {
		return err
	}
	for i, h := range p.Hosts {
		if h.Taskbook != "" {
			err := h.ParseTaskbook()
			if err != nil {
				return err
			}
			p.Hosts[i] = h
		}
	}

	return nil
}

// Play executes the palybook.
func (p *Playbook) Play() []Result {
	maxCh := make(chan bool, p.MaxNum)
	chs := make([]chan Result, len(p.Hosts))

	for i, host := range p.Hosts {
		chs[i] = make(chan Result, 1)
		maxCh <- true

		go func(chLimit chan bool, ch chan Result, host Host) {
			Do(
				host.User,
				host.Password,
				host.IP,
				host.CertKey,
				host.Port,
				host.Timeout,
				host.Ciphers,
				host.Tasks,
				host.Platform,
				ch)

			<-chLimit
		}(maxCh, chs[i], host)
	}

	results := []Result{}

	for _, ch := range chs {
		r := <-ch
		if r.Host != "" {
			results = append(results, r)
		}
	}

	return results
}

// StreamOut show the results.
func (p *Playbook) StreamOut() {
	return
}
