package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	//"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/def"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

var (
	singleQuotesRegex  = regexp.MustCompile(`\A'(.*)'\z`)
	doubleQuotesRegex  = regexp.MustCompile(`\A"(.*)"\z`)
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeCharsRegex = regexp.MustCompile(`\\([^$])`)
)

func (p *Project) getAppShellCommand(d *def.App) (ShellCommand, error) {
	// get app brew service
	serviceList, err := LoadServiceList()
	if err != nil {
		return ShellCommand{}, err
	}
	brewAppService, err := serviceList.MatchDef(d)
	if err != nil {
		return ShellCommand{}, err
	}
	brewServiceList := make([]*Service, 0)
	brewServiceList = append(brewServiceList, brewAppService)
	for _, service := range p.Services {
		brewService, err := serviceList.MatchDef(&service)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				continue
			}
			return ShellCommand{}, err
		}
		brewServiceList = append(brewServiceList, brewService)
	}
	// generate pathes
	envPaths := []string{
		filepath.Join(p.Path, ".global", "bin"),
		filepath.Join(p.Path, ".global", "vendor", "bin"),
		filepath.Join(p.Path, ".global", "node_modules", "bin"),
		filepath.Join(p.Path, ".platformsh", "bin"),
		filepath.Join(GetDir(HomeDir), ".pyenv", "versions", "3.10.0", "bin"),
		filepath.Join(GetDir(HomeDir), ".pyenv", "versions", "2.7.18", "bin"),
	}
	// inject env vars
	env := make([]string, 0)
	env = append(env, ServicesEnv(brewServiceList)...)

	//env = append(env, "HOME="+p.Path)
	//env = append(env, fmt.Sprintf("NVM_DIR=%s/.nvm", GetDir(HomeDir)))
	env = append(env, fmt.Sprintf("TERM=%s", os.Getenv("TERM")))

	for k, v := range p.Env(d) {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	envKeyMap := make(map[string]int)

	for k, v := range env {

		envKeyMap[strings.Split(v, "=")[0]] = 1

		if strings.HasPrefix(v, "PATH") {
			env[k] = "PATH=" + strings.Join(envPaths, ":") + ":" + strings.TrimPrefix(env[k], "PATH=")
		}
	}

	// run interactive shell
	cmd := NewShellCommand()

	// load .env.prod file
	if _, nofile_err := os.Stat(".env.prod"); nofile_err == nil {

		output.WriteStdout(" loading .env.prod \n")

		if envMap, err := readFile(".env.prod");err == nil {

			for key, value := range envMap {
				_, ok := envKeyMap[key]
				if !ok || key == "PLATFORM_VARIABLES" {
					env = append(env, fmt.Sprintf("%s=%s", key,  strings.ReplaceAll(value, "'", "\"")))
				}
			}
			
		} else {
			output.WriteStdout(" Failed parsing .env.prod \n")
			return cmd, err
		}

	} else {
		output.WriteStdout(" No .env.prod file to process \n")
	}

	cmd.Args = []string{}
	cmd.Env = env

	return cmd, nil
}

// Shell opens a shell in given app context.
func (p *Project) Shell(d *def.App) error {
	output.Info(fmt.Sprintf("Access shell for %s.", d.Name))
	cmd, err := p.getAppShellCommand(d)
	if err != nil {
		return err
	}
	if err := cmd.Drop(); err != nil {
		return err
	}
	return nil
}

// Command executes a shell command in given app context.
func (p *Project) Command(d *def.App, cmdStr string) error {
	output.LogInfo(fmt.Sprintf("Run command '%s' in '%s'.", cmdStr, d.Name))
	cmd, err := p.getAppShellCommand(d)
	if err != nil {
		return err
	}
	cmdStr = "source $(brew --prefix nvm)/nvm.sh && " + cmdStr
	cmd.Args = []string{"-c", cmdStr}
	if err := cmd.Interactive(); err != nil {
		return err
	}
	return nil
}

func readFile(filename string) (envMap map[string]string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return Parse(file)
}

// Parse reads an env file from io.Reader, returning a map of keys and values.
func Parse(r io.Reader) (envMap map[string]string, err error) {
	envMap = make(map[string]string)

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return
	}

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			var key, value string
			key, value, err = parseLine(fullLine, envMap)

			if err != nil {
				return
			}
			envMap[key] = value
		}
	}
	return
}

func parseLine(line string, envMap map[string]string) (key string, value string, err error) {
	if len(line) == 0 {
		err = errors.New("zero length string")
		return
	}

	// ditch the comments (but keep quoted hashes)
	if strings.Contains(line, "#") {
		segmentsBetweenHashes := strings.Split(line, "#")
		quotesAreOpen := false
		var segmentsToKeep []string
		for _, segment := range segmentsBetweenHashes {
			if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, segment)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, segment)
			}
		}

		line = strings.Join(segmentsToKeep, "#")
	}

	firstEquals := strings.Index(line, "=")
	firstColon := strings.Index(line, ":")
	splitString := strings.SplitN(line, "=", 2)
	if firstColon != -1 && (firstColon < firstEquals || firstEquals == -1) {
		//this is a yaml-style line
		splitString = strings.SplitN(line, ":", 2)
	}

	if len(splitString) != 2 {
		err = errors.New("Can't separate key from value")
		return
	}

	// Parse the key
	key = splitString[0]
	if strings.HasPrefix(key, "export") {
		key = strings.TrimPrefix(key, "export")
	}
	key = strings.TrimSpace(key)

	key = exportRegex.ReplaceAllString(splitString[0], "$1")

	// Parse the value
	value = parseValue(splitString[1], envMap)
	return
}

func isIgnoredLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}

var exportRegex = regexp.MustCompile(`^\s*(?:export\s+)?(.*?)\s*$`)

func parseValue(value string, envMap map[string]string) string {

	// trim
	value = strings.Trim(value, " ")

	// check if we've got quoted values or possible escapes
	if len(value) > 1 {
		singleQuotes := singleQuotesRegex.FindStringSubmatch(value)

		doubleQuotes := doubleQuotesRegex.FindStringSubmatch(value)

		if singleQuotes != nil || doubleQuotes != nil {
			// pull the quotes off the edges
			value = value[1 : len(value)-1]
		}

		if doubleQuotes != nil {
			// expand newlines
			value = escapeRegex.ReplaceAllStringFunc(value, func(match string) string {
				c := strings.TrimPrefix(match, `\`)
				switch c {
				case "n":
					return "\n"
				case "r":
					return "\r"
				default:
					return match
				}
			})
			// unescape characters
			value = unescapeCharsRegex.ReplaceAllString(value, "$1")
		}

		if singleQuotes == nil {
			value = expandVariables(value, envMap)
		}
	}

	return value
}

var expandVarRegex = regexp.MustCompile(`(\\)?(\$)(\()?\{?([A-Z0-9_]+)?\}?`)

func expandVariables(v string, m map[string]string) string {
	return expandVarRegex.ReplaceAllStringFunc(v, func(s string) string {
		submatch := expandVarRegex.FindStringSubmatch(s)

		if submatch == nil {
			return s
		}
		if submatch[1] == "\\" || submatch[2] == "(" {
			return submatch[0][1:]
		} else if submatch[4] != "" {
			return m[submatch[4]]
		}
		return s
	})
}