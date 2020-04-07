# ppsh -- Pull/Push via SSH

PPSH is a [Ansible](https://github.com/ansible/ansible)-like Tool and Library written in Go. Ansible is a radically simple IT automation system, and it is Awesome, but it is written in Python and needs Python installed and sometime annoyed configurations, so here comes PPSH. You just need the precompile bin file and a well defined [YAML](http://www.yaml.org/spec/1.2/spec.html) file (see more examples in [app/file](https://github.com/haotrr/ppsh/tree/master/app/file))  to run it, and for simple tasks you even just run it with the well defined arguments. PPSH is also a library and can be easily integrated into your application. More detail will come soon...

## Build
```bash
make
```

## App Usage
```bash
$ ./ppsh help
NAME:
   ppsh - Pull or Push via SSH in your cluster hosts.

USAGE:
   ppsh [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --hosts HOST[;HOST], -H HOST[;HOST]                       host list, in the form of HOST[;HOST]
   --cmds CMD[;CMD], -c CMD[;CMD]                            command list, in the form of CMD[;CMD]
   --ciphers CIPHER[;CIPHER], -C CIPHER[;CIPHER]             cipher list, in the form of CIPHER[;CIPHER]
   --ip-range IP-IP[;{IP-IP|IP/XX}, -I IP-IP[;{IP-IP|IP/XX}  ip range, in the form of IP-IP[;{IP-IP|IP/XX}
   --user USER, -u USER                                      ssh login USER (default: "root")
   --password PASSWORD, -w PASSWORD                          ssh login PASSWORD
   --cert-key FILE, -k FILE                                  ssh private key FILE
   --playbook FILE, -p FILE                                  load playbook from path FILE
   --taskbook FILE, -t FILE                                  load taskbook from path FILE
   --format PLAIN|JSON, -f PLAIN|JSON                        output as PLAIN|JSON (default: "plain")
   --platform LINUX|OTHER, -S LINUX|OTHER                    platform as LINUX|OTHER (default: "linux")
   --output STDOUT|FILE, -o STDOUT|FILE                      output to STDOUT|FILE (default: "stdout")
   --timeout TIMEOUT, -s TIMEOUT                             TIMEOUT in second (default: 30)
   --port PORT, -P PORT                                      ssh PORT (default: 22)
   --max-run-count COUNT, -n COUNT                           max runing COUNT (default: 20)
   --help, -h                                                show help
   --version, -v                                             print the version
```
See more examples in [app/file/test.txt](https://github.com/haotrr/ppsh/blob/master/app/file/test.txt).

## As Library
See more details in [godoc](https://godoc.org/github.com/haotrr/ppsh).

## Roadmap
- [ ] Upload files
- [ ] Download files

## License
MIT License, see detail in [LICENSE](https://github.com/haotrr/ppsh/blob/master/LICENSE).