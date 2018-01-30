package ppsh

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Result struct {
	UUID    string `json:"uuid,omitempty"`
	Host    string `json:"host"`
	Cmd     string `json:"cmd"`
	Success bool   `json:"success"`
	Code    int    `json:"code,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Host struct {
	UUID     string   `yaml:"-"`
	IP       string   `yaml:"ip"`
	Port     int      `yaml:"port"`
	User     string   `yaml:"user"`
	Password string   `yaml:"password"`
	CertKey  string   `yaml:"cert-key"`
	Ciphers  []string `yaml:"ciphers"`
	Tasks    []string `yaml:"tasks"`
	Taskbook string   `yaml:"taskbook"`
	Result   Result   `yaml:"-"`
	Platform string   `yaml:"platform"`
	Timeout  int      `yaml:"timeout"`
}

func (h *Host) ParseTaskbook() error {
	f, err := os.Open(h.Taskbook)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &h.Tasks)
	if err != nil {
		return err
	}

	return nil
}

type Platform string

const (
	LINUX = "linux"
	OTHER = "other"
)

type Format int

const (
	JSON Format = iota + 1
	PLAIN
)
