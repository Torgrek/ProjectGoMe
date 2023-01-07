package main

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
)

func discordModule() {

	var tokens []tokensselection = getBotTokens()

	var id int = 0

	for _, token := range tokens {
		var distoken string = "Bot " + token.token
		discord, errdiscrod := discordgo.New(distoken)
		checkIfNil(errdiscrod)
		target := &tokens[id]
		target.connection = discord
		target.getConnectionDiscord()
		id++
	}

	initializeAllLogicToConnections(tokens)

}

type voicesessions struct {
	voicechannel *discordgo.VoiceConnection
	session      *discordgo.Session
	channelId    string
	userId       string
	order        int
	bottype      int
	locked       bool
}

func initializeAllLogicToConnections(listOfConnections []tokensselection) {

	for _, token := range listOfConnections {

		botname := token.name
		bottype := token.bottype
		connection := token.connection
		iterator := 0

		switch token.bottype {
		case 1:
			fmt.Println("INITIALISING " + botname)
			connection.AddHandler(ReadyHandler)
			connection.AddHandler(UpdateVoiceChannelEventToHand)
			connection.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates
			err := connection.Open()
			if checkIfNil(err) {
				return
			}
			createVoiceSession(connection, iterator, token.bottype)
			fmt.Println("INITIALISED " + botname)
			iterator++
		default:
			fmt.Println("TYPE NOT FOUND " + strconv.Itoa(bottype) + " FOR " + botname)
		}
	}

}

func getFreeBot(bottype int) *discordgo.Session {

	var listofTypesCorrectlyBots []voicesessions
	for _, element := range voiceSessionMaster {
		if element.bottype == bottype && !element.locked {
			listofTypesCorrectlyBots = append(listofTypesCorrectlyBots, element)
		}
	}

	var indexOfMin int = 999999
	var indexOfMinOrder int = 999999
	for i := 0; i < len(listofTypesCorrectlyBots); i++ {
		currentElement := listofTypesCorrectlyBots[i]
		if currentElement.order < indexOfMinOrder {
			indexOfMinOrder = currentElement.order
			indexOfMin = i
		}
	}

	if indexOfMin == 999999 {
		return nil
	}
	return listofTypesCorrectlyBots[indexOfMin].session
}

type guildSlice struct {
	guildid   string
	guildname string
}

type guildUsersSlice struct {
	userid string
}

type guildMember struct {
	userid  string
	guildid string
}

type guildRolesSlice struct {
	roleid  string
	guildid string
}

func ReadyHandler(session *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Updating guilds information...")

	//++ Check guilds
	currentGuilds := event.Guilds
	var GuildsToAdd []guildSlice

	lenGuilds := len(currentGuilds)

	if lenGuilds == 0 {
		fmt.Println("Guilds not found")
		return
	}

	for i := 0; i < lenGuilds; i++ {
		sliceToAdd := guildSlice{}
		sliceToAdd.guildid = currentGuilds[i].ID
		sliceToAdd.guildname = currentGuilds[i].Name
		GuildsToAdd = append(GuildsToAdd, sliceToAdd)
	}

	currentDriver := globalruntimeparams.driver
	rows, err := currentDriver.Query("select * from guilds")

	checkIfNil(err)

	for rows.Next() {
		stringToCheck := guildsSelection{}
		err = rows.Scan(&stringToCheck.guildid)
		if checkIfNil(err) {
			continue
		}

		stringedGuildID := strconv.FormatInt(stringToCheck.guildid, 10)

		idToDelete := 0

		if lenGuilds == 0 {
			lenGuilds = len(GuildsToAdd)
			continue
		}

		for i := 0; i < lenGuilds; i++ {

			if GuildsToAdd[i].guildid == stringedGuildID {
				idToDelete = i - 1
				break
			}
		}

		GuildsToAdd = append(GuildsToAdd[:idToDelete], GuildsToAdd[idToDelete+1:]...)
		lenGuilds = len(GuildsToAdd)
	}

	if lenGuilds > 0 {

		var queryString string = "INSERT INTO guilds (guildid) VALUES "

		for _, elementOfResult := range GuildsToAdd {
			queryString += "(" + elementOfResult.guildid + "),"
		}

		queryString = queryString[0 : len(queryString)-1]
		_, errstmt := currentDriver.Query(queryString)
		checkIfNil(errstmt)
		fmt.Println("Guild updated...")
	} else {
		fmt.Println("Nothing to update...")
	}
	//-- Check guilds

	fmt.Println("Updating users...")

	UsersRolesArray := getRolesFromDB()

	UsersIdArray := getUsersFromDB()

	var ArrayToCheckMembers []guildMember

	for i := 0; i < len(currentGuilds); i++ {
		foundedMembers, memberserr := session.GuildMembers(currentGuilds[i].ID, "000000000000000000", 1000)
		checkIfNil(memberserr)
		for _, member := range foundedMembers {
			RolesToCheck := member.Roles
			isRoleFounded := false
			for _, role := range RolesToCheck {
				for _, roleInDB := range UsersRolesArray {
					if role == roleInDB.roleid {
						isRoleFounded = true
					}
				}

				if isRoleFounded {
					break
				}
			}
			if isRoleFounded {
				userIdToCheck := member.User.ID
				isUserIdNotFound := true
				for _, userid := range UsersIdArray {
					if userid == userIdToCheck {
						isUserIdNotFound = false
					}
				}

				if isUserIdNotFound {
					var guildMemberToAdd guildMember
					guildMemberToAdd.userid = userIdToCheck
					guildMemberToAdd.guildid = currentGuilds[i].ID
					ArrayToCheckMembers = append(ArrayToCheckMembers, guildMemberToAdd)
				}
			}
		}

	}

	if len(ArrayToCheckMembers) > 0 {

		var queryString string = "INSERT INTO users (userid) VALUES "

		for _, elementOfResult := range ArrayToCheckMembers {
			queryString += "(" + elementOfResult.userid + "),"
		}

		queryString = queryString[0 : len(queryString)-1]
		_, errstmt := currentDriver.Query(queryString)
		checkIfNil(errstmt)
		fmt.Println("Users updated...")
	} else {
		fmt.Println("Nothing to update...")
	}

}

type guildsSelection struct {
	guildid int64
}

func UpdateVoiceChannelEventToHand(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {

	var VoiceStateCurrent *discordgo.VoiceState = event.VoiceState

	var UserID string = VoiceStateCurrent.UserID
	//Return if author is bot
	if UserID == session.State.User.ID || event.VoiceState.Member.User.Bot {
		return
	}

	var ChannelID string = VoiceStateCurrent.ChannelID
	if ChannelID == "" {
		clearVoiceSession(event.BeforeUpdate.ChannelID, event.BeforeUpdate.UserID)
		return
	}

	if event.BeforeUpdate != nil {
		if event.BeforeUpdate.ChannelID != ChannelID {
			clearVoiceSession(event.BeforeUpdate.ChannelID, event.BeforeUpdate.UserID)
		}
	}

	if session != getFreeBot(1) {
		return
	}

	var GuildID string = VoiceStateCurrent.GuildID
	//var bytesToRead := getUserAviliableSound(GuildID, UserID)

	vc, err := session.ChannelVoiceJoin(GuildID, ChannelID, false, true)
	if err != nil {
		for _, element := range session.VoiceConnections {
			if element.ChannelID == ChannelID {
				vc = element
			}
		}
		updateVoiceSession(session, vc, ChannelID, UserID)
		clearVoiceSession(ChannelID, UserID)
		return
	}
	updateVoiceSession(session, vc, ChannelID, UserID)
	fmt.Println(vc)

}

func createVoiceSession(session *discordgo.Session, order int, bottype int) {

	var currentStructure voicesessions
	currentStructure.session = session
	currentStructure.order = order
	currentStructure.locked = false
	currentStructure.bottype = bottype
	voiceSessionMaster = append(voiceSessionMaster, currentStructure)

}

func updateVoiceSession(session *discordgo.Session, voiceChannel *discordgo.VoiceConnection, channelId string, userID string) {

	iterator := 0
	for _, sessionVoice := range voiceSessionMaster {
		if sessionVoice.session == session {
			voiceSessionMaster[iterator].voicechannel = voiceChannel
			voiceSessionMaster[iterator].channelId = channelId
			voiceSessionMaster[iterator].userId = userID
			voiceSessionMaster[iterator].locked = true
		}
		iterator++
	}
}

func clearVoiceSession(previousChannelId string, UserID string) {
	iterator := 0
	for _, sessionVoice := range voiceSessionMaster {
		if previousChannelId == sessionVoice.channelId && UserID == sessionVoice.userId {
			if voiceSessionMaster[iterator].voicechannel != nil {
				voiceSessionMaster[iterator].voicechannel.Disconnect()
			}
			voiceSessionMaster[iterator].channelId = ""
			voiceSessionMaster[iterator].userId = ""
			voiceSessionMaster[iterator].locked = false
			break
		}
		iterator++
	}

}

func getRolesFromDB() []guildRolesSlice {
	//++ Get roles
	currentDriver := globalruntimeparams.driver
	rowsroles, errroles := currentDriver.Query("select * from roles")
	checkIfNil(errroles)

	var UsersRolesArray []guildRolesSlice

	for rowsroles.Next() {
		userStruct := guildRolesSlice{}
		err := rowsroles.Scan(&userStruct.roleid, &userStruct.guildid)
		if checkIfNil(err) {
			continue
		}
		UsersRolesArray = append(UsersRolesArray, userStruct)
	}
	//-- Get roles

	return UsersRolesArray
}

func getUsersFromDB() []string {
	//++ Get users
	currentDriver := globalruntimeparams.driver
	rowsusrs, errusrs := currentDriver.Query("select * from users")
	checkIfNil(errusrs)

	var UsersIdArray []string

	for rowsusrs.Next() {
		userStruct := guildUsersSlice{}
		err := rowsusrs.Scan(&userStruct.userid)
		if checkIfNil(err) {
			continue
		}
		UsersIdArray = append(UsersIdArray, userStruct.userid)
	}
	//-- Get users

	return UsersIdArray
}

type tokensselection struct {
	token      string
	name       string
	connection *discordgo.Session
	bottype    int
}

func (token tokensselection) getConnectionDiscord() bool {
	var status string
	var result bool = false
	if token.connection != nil {
		status = "OK"
		result = true
	} else {
		status = "NOT CONNECTED"
	}

	connectiontype := strconv.Itoa(token.bottype)

	fmt.Println(token.name + " " + connectiontype + " " + status)

	return result
}

func getBotTokens() []tokensselection {
	rows, err := globalruntimeparams.driver.Query("select * from sessionsmaster")

	checkIfNil(err)

	var result []tokensselection

	for rows.Next() {
		stringToAdd := tokensselection{}
		err = rows.Scan(&stringToAdd.token, &stringToAdd.name, &stringToAdd.bottype)
		if checkIfNil(err) {
			continue
		}
		result = append(result, stringToAdd)
	}

	return result

}
