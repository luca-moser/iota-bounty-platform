package misc

import (
	"crypto/rand"
	"fmt"
	"github.com/iotaledger/iota.go/consts"
	"github.com/mattn/go-colorable"
	"github.com/pkg/errors"
	"gopkg.in/inconshreveable/log15.v2"
	"net/url"
	"os"
	"strings"
)

var Debug = false

func init() {
	os.Mkdir("./logs", 0777)
}

func GetLogger(name string) (log15.Logger, error) {

	// open a new logfile
	fileHandler, err := log15.FileHandler(fmt.Sprintf("./logs/%s.log", name), log15.LogfmtFormat())
	if err != nil {
		return nil, err
	}

	handler := log15.MultiHandler(
		fileHandler,
		log15.StreamHandler(colorable.NewColorableStdout(), log15.TerminalFormat()),
	)
	if !Debug {
		handler = log15.LvlFilterHandler(log15.LvlInfo, handler)
	}
	logger := log15.New("comp", name)
	logger.SetHandler(handler)
	return logger, nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHJIKLMNOPQRSTUVWXYZ123456789"

func GenerateAlphanumericCode(length int) string {
	by := make([]byte, length)
	if _, err := rand.Read(by); err != nil {
		panic(err)
	}
	var pw string
	for _, b := range by {
		pw += string(letters[int(b)%len(letters)])
	}
	return pw
}

const seedLength = 81

var tryteAlphabetLength = byte(len(consts.TryteAlphabet))

func GenerateSeed() (string, error) {
	var by [seedLength]byte
	if _, err := rand.Read(by[:]); err != nil {
		return "", err
	}
	var seed string
	for _, b := range by {
		seed += string(consts.TryteAlphabet[b%tryteAlphabetLength])
	}
	return seed, nil
}

const githubFrag = "github.com"

var ErrRepoURLInvalid = errors.New("repository URL invalid")

func ExtractOwnerAndNameFromGitHubURL(repoURL string) (string, string, error) {
	_, err := url.Parse(repoURL)
	if err != nil {
		return "", "", ErrRepoURLInvalid
	}
	if !strings.Contains(repoURL, githubFrag) {
		return "", "", ErrRepoURLInvalid
	}
	urlSplit := strings.Split(repoURL, "/")
	splitLength := len(urlSplit)
	if splitLength < 2 {
		return "", "", ErrRepoURLInvalid
	}
	return urlSplit[splitLength-2], urlSplit[splitLength-1], nil
}
