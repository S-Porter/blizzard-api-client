package wow

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ApiClient struct {
	Host      string
	Locale    string
	Secret    string
	PublicKey string
}

var apiClient *ApiClient = nil

func CurrentApiClient() *ApiClient {
	return apiClient
}

// NewApiClient accepts a region (US, EU, KR, TW, ZH) and an optional
// associated locale to return a new instance of ApiClient. If the
// locale is an empty string, the default locale for that region will
// be used.
func NewApiClient(region string, locale string) (*ApiClient, error) {
	var host string
	var validLocales []string
	switch region {
	case "US", "United States":
		host = "us.api.battle.net"
		validLocales = []string{"en_US", "es_MX", "pt_BR"}
	case "EU", "Europe":
		host = "eu.battle.net"
		validLocales = []string{"en_GB", "es_ES", "fr_FR", "ru_RU", "de_DE", "pt_PT", "it_IT"}
	case "KR", "Korea":
		host = "kr.battle.net"
		validLocales = []string{"ko_KR"}
	case "TW", "Taiwan":
		host = "tw.battle.net"
		validLocales = []string{"zh_TW"}
	case "ZH", "CN", "China":
		host = "www.battle.com.cn"
		validLocales = []string{"zh_CN"}
	default:
		return nil, errors.New(fmt.Sprintf("Region '%s' is not valid", region))
	}

	var client *ApiClient
	if locale == "" {
		client = &ApiClient{Host: host, Locale: validLocales[0]}
	} else {
		for _, valid := range validLocales {
			if valid == locale {
				client = &ApiClient{Host: host, Locale: locale}
			}
		}
	}
	if client != nil {
		apiClient = client
		return client, nil
	}

	return nil, errors.New(fmt.Sprintf("Locale '%s' is not valid for region '%s'", locale, region))
}

func (a *ApiClient) GetAchievement(id int) (*Achievement, error) {
	jsonBlob, err := a.get(fmt.Sprintf("achievement/%d", id))
	if err != nil {
		return nil, err
	}
	achieve := &Achievement{}
	err = json.Unmarshal(jsonBlob, achieve)
	if err != nil {
		return nil, err
	}
	return achieve, nil
}

func (a *ApiClient) GetAuctionData(realm string) (*AuctionData, error) {
	jsonBlob, err := a.get(fmt.Sprintf("auction/data/%s", realm))
	if err != nil {
		return nil, err
	}
	auctionData := &AuctionData{}
	err = json.Unmarshal(jsonBlob, auctionData)
	if err != nil {
		return nil, err
	}
	return auctionData, nil
}

func (a *ApiClient) GetBattlePetAbility(id int) (*BattlePetAbility, error) {
	jsonBlob, err := a.get(fmt.Sprintf("battlePet/ability/%d", id))
	if err != nil {
		return nil, err
	}
	ability := &BattlePetAbility{}
	err = json.Unmarshal(jsonBlob, ability)
	if err != nil {
		return nil, err
	}
	return ability, nil
}

func (a *ApiClient) GetBattlePetSpecies(id int) (*BattlePetSpecies, error) {
	jsonBlob, err := a.get(fmt.Sprintf("battlePet/species/%d", id))
	if err != nil {
		return nil, err
	}
	species := &BattlePetSpecies{}
	err = json.Unmarshal(jsonBlob, species)
	if err != nil {
		return nil, err
	}
	return species, nil
}

func (a *ApiClient) GetBattlePet(id int, level int, breedId int, qualityId int) (*BattlePet, error) {
	jsonBlob, err := a.getWithParams(
		fmt.Sprintf("battlePet/stats/%d", id),
		map[string]string{
			"level":     strconv.Itoa(level),
			"breedId":   strconv.Itoa(breedId),
			"qualityId": strconv.Itoa(qualityId),
		})
	if err != nil {
		return nil, err
	}

	pet := &BattlePet{}
	err = json.Unmarshal(jsonBlob, pet)
	if err != nil {
		return nil, err
	}
	return pet, nil
}

func (a *ApiClient) GetBattlePetStats(id int, level int, breedId int, qualityId int) (*BattlePet, error) {
	return a.GetBattlePet(id, level, breedId, qualityId)
}

// Will return the ApiClient's region's challenges if realm is empty
// string.
func (a *ApiClient) GetChallenges(realm string) ([]*Challenge, error) {
	if realm == "" {
		realm = "region"
	}
	jsonBlob, err := a.get(fmt.Sprintf("challenge/%s", realm))
	if err != nil {
		return nil, err
	}
	challengeSet := &challengeList{}
	err = json.Unmarshal(jsonBlob, challengeSet)
	if err != nil {
		return nil, err
	}
	return challengeSet.Challenges, nil
}

func (a *ApiClient) GetChallenge(realm string) ([]*Challenge, error) {
	return a.GetChallenges(realm)
}

func (a *ApiClient) GetCharacter(realm string, characterName string) (*Character, error) {
	return a.GetCharacterWithFields(realm, characterName, make([]string, 0))
}

func (a *ApiClient) GetCharacterWithFields(realm string, characterName string, fields []string) (*Character, error) {
	err := validateCharacterFields(fields)
	if err != nil {
		return nil, err
	}
	jsonBlob, err := a.getWithParams(fmt.Sprintf("character/%s/%s", realm, characterName), map[string]string{"fields": strings.Join(fields, ",")})

	if err != nil {
		return nil, err
	}
	char := NewCharacter(a)
	err = json.Unmarshal(jsonBlob, char)
	if err != nil {
		return nil, err
	}
	return char, nil
}

func (a *ApiClient) GetItem(id int) (*Item, error) {
	jsonBlob, err := a.get(fmt.Sprintf("item/%d", id))
	if err != nil {
		return nil, err
	}
	item, err := NewItemFromJson(jsonBlob)
	if err != nil {
		return nil, err
	}

	return item, err
}

func (a *ApiClient) GetItemSet(id int) (*ItemSet, error) {
	jsonBlob, err := a.get(fmt.Sprintf("item/set/%d", id))
	if err != nil {
		return nil, err
	}
	itemSet := &ItemSet{}
	err = json.Unmarshal(jsonBlob, itemSet)
	if err != nil {
		return nil, err
	}

	return itemSet, err
}

func (a *ApiClient) GetGuild(realm string, guildName string) (*Guild, error) {
	return a.GetGuildWithFields(realm, guildName, make([]string, 0))
}

func (a *ApiClient) GetGuildWithFields(realm string, guildName string, fields []string) (*Guild, error) {
	err := validateGuildFields(fields)
	if err != nil {
		return nil, err
	}
	jsonBlob, err := a.getWithParams(fmt.Sprintf("guild/%s/%s", realm, url.QueryEscape(guildName)), map[string]string{"fields": strings.Join(fields, ",")})
	if err != nil {
		return nil, err
	}
	guild := &Guild{}
	err = json.Unmarshal(jsonBlob, guild)
	if err != nil {
		return nil, err
	}
	return guild, nil
}

func (a *ApiClient) GetPvPLeaderboard(bracket string) ([]*PvPLeaderboardRow, error) {
	jsonBlob, err := a.get(fmt.Sprintf("leaderboard/%s", bracket))

	leaderboard := &pvpLeaderboard{}
	err = json.Unmarshal(jsonBlob, leaderboard)
	if err != nil {
		return nil, err
	}
	return leaderboard.Rows, nil
}

func (a *ApiClient) GetQuest(id int) (*Quest, error) {
	jsonBlob, err := a.get(fmt.Sprintf("quest/%d", id))

	quest := &Quest{}
	err = json.Unmarshal(jsonBlob, quest)
	if err != nil {
		return nil, err
	}
	return quest, nil
}

func (a *ApiClient) GetRealmStatus() ([]*RealmStatus, error) {
	jsonBlob, err := a.get("realm/status")

	list := &realmStatusList{}
	err = json.Unmarshal(jsonBlob, list)
	if err != nil {
		return nil, err
	}
	return list.Realms, nil
}

func (a *ApiClient) GetRecipe(id int) (*Recipe, error) {
	jsonBlob, err := a.get(fmt.Sprintf("recipe/%d", id))

	recipe := &Recipe{}
	err = json.Unmarshal(jsonBlob, recipe)
	if err != nil {
		return nil, err
	}
	return recipe, nil
}

func (a *ApiClient) GetSpell(id int) (*Spell, error) {
	jsonBlob, err := a.get(fmt.Sprintf("spell/%d", id))

	spell := &Spell{}
	err = json.Unmarshal(jsonBlob, spell)
	if err != nil {
		return nil, err
	}
	return spell, nil
}

func (a *ApiClient) GetBattlegroups() ([]*Battlegroup, error) {
	jsonBlob, err := a.get("data/battlegroups/")

	battlegroupList := &battlegroupList{}
	err = json.Unmarshal(jsonBlob, battlegroupList)
	if err != nil {
		return nil, err
	}
	return battlegroupList.Battlegroups, nil
}

func (a *ApiClient) GetRaces() ([]*Race, error) {
	jsonBlob, err := a.get("data/character/races")

	raceList := &raceList{}
	err = json.Unmarshal(jsonBlob, raceList)
	if err != nil {
		return nil, err
	}
	return raceList.Races, nil
}

func (a *ApiClient) GetClasses() ([]*Class, error) {
	jsonBlob, err := a.get("data/character/classes")

	classList := &classList{}
	err = json.Unmarshal(jsonBlob, classList)
	if err != nil {
		return nil, err
	}
	return classList.Classes, nil
}

func (a *ApiClient) GetAchievements() ([]*Achievement, error) {
	jsonBlob, err := a.get("data/character/achievements")

	achievementList := &achievementData{}
	err = json.Unmarshal(jsonBlob, achievementList)
	if err != nil {
		return nil, err
	}
	return achievementList.Achievements, nil
}

func (a *ApiClient) GetGuildRewards() ([]*GuildReward, error) {
	jsonBlob, err := a.get("data/guild/rewards")

	guildRewardList := &guildRewardList{}
	err = json.Unmarshal(jsonBlob, guildRewardList)
	if err != nil {
		return nil, err
	}
	return guildRewardList.Rewards, nil
}

func (a *ApiClient) GetGuildPerks() ([]*GuildPerk, error) {
	jsonBlob, err := a.get("data/guild/perks")

	guildPerkList := &guildPerkList{}
	err = json.Unmarshal(jsonBlob, guildPerkList)
	if err != nil {
		return nil, err
	}
	return guildPerkList.Perks, nil
}

func (a *ApiClient) GetGuildAchievements() ([]*Achievement, error) {
	jsonBlob, err := a.get("data/guild/achievements")

	guildAchievementList := &guildAchievementList{}
	err = json.Unmarshal(jsonBlob, guildAchievementList)
	if err != nil {
		return nil, err
	}
	return guildAchievementList.Achievements, nil
}

func (a *ApiClient) GetItemClasses() ([]*ItemClass, error) {
	jsonBlob, err := a.get("data/item/classes")

	itemClassList := &itemClassList{}
	err = json.Unmarshal(jsonBlob, itemClassList)
	if err != nil {
		return nil, err
	}
	return itemClassList.Classes, nil
}

func (a *ApiClient) GetTalents() (*ClassTalentList, error) {
	jsonBlob, err := a.get("data/talents")

	talents := &ClassTalentList{}
	err = json.Unmarshal(jsonBlob, talents)
	if err != nil {
		return nil, err
	}
	return talents, nil
}

func (a *ApiClient) GetPetTypes() ([]*PetType, error) {
	jsonBlob, err := a.get("data/pet/types")

	petTypes := &petTypeList{}
	err = json.Unmarshal(jsonBlob, petTypes)
	if err != nil {
		return nil, err
	}
	return petTypes.PetTypes, nil
}

func validateGuildFields(fields []string) error {
	validFields := []string{
		"members",
		"achievements",
		"news",
		"challenge"}
	return validateFields(validFields, fields)
}

func validateCharacterFields(fields []string) error {
	validFields := []string{
		"achievements",
		"appearance",
		"feed",
		"guild",
		"hunterPets",
		"items",
		"mounts",
		"pets",
		"petSlots",
		"professions",
		"progression",
		"pvp",
		"quests",
		"reputation",
		"stats",
		"talents",
		"titles"}
	return validateFields(validFields, fields)
}

func validateFields(validFields []string, fields []string) error {
	badFields := make([]string, 0)
	var exists bool
	for _, field := range fields {
		exists = false
		for _, valid := range validFields {
			if valid == field {
				exists = true
			}
		}
		if !exists {
			badFields = append(badFields, field)
		}
	}
	if len(badFields) != 0 {
		return errors.New(fmt.Sprintf("The following fields are not valid: %v", badFields))
	} else {
		return nil
	}
}

func (a *ApiClient) get(path string) ([]byte, error) {
	return a.getWithParams(path, make(map[string]string))
}

func (a *ApiClient) getWithParams(path string, queryParams map[string]string) ([]byte, error) {
	client := &http.Client{}
	var url *url.URL
	var request *http.Request
	var err error

	if len(a.Secret) > 0 {
		url = a.url(path, queryParams, true)
		request, err = http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return make([]byte, 0), err
		}
	} else {
		url = a.url(path, queryParams, false)
		request, err = http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return make([]byte, 0), err
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return make([]byte, 0), err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return body, nil
}

func (a *ApiClient) url(path string, queryParamPairs map[string]string, ssl bool) *url.URL {
	queryParamPairs["locale"] = a.Locale
	queryParamPairs["apikey"] = a.Secret
	queryParamList := make([]string, 0)
	for k, v := range queryParamPairs {
		queryParamList = append(queryParamList, k+"="+v)
	}
	var scheme string
	if ssl {
		scheme = "https"
	} else {
		scheme = "http"
	}
	return &url.URL{
		Scheme:   scheme,
		Host:     a.Host,
		Path:     "/wow/" + path,
		RawQuery: strings.Join(queryParamList, "&"),
	}
}

func (a *ApiClient) authorizationString(signature string) string {
	return fmt.Sprintf(" BNET %s:%s", a.PublicKey, signature)
}

func (a *ApiClient) signature(verb string, path string) string {
	url := a.url(path, make(map[string]string), true)
	toBeSigned := []byte(strings.Join([]string{verb, time.Now().String(), url.Path, ""}, "\n"))
	mac := hmac.New(sha1.New, []byte(a.Secret))
	_, err := mac.Write(toBeSigned)
	if err != nil {
		handleError(err)
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func handleError(err error) {
	panic(err)
}
