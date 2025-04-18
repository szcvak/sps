package network

import (
	"github.com/szcvak/sps/pkg/messages"
	"github.com/szcvak/sps/pkg/messaging"
)

var ClientRegistry = make(map[uint16]func() messaging.ClientMessage)

func registerClientMessage(id uint16, factory func() messaging.ClientMessage) {
	ClientRegistry[id] = factory
}

func init() {
	registerClientMessage(10101, func() messaging.ClientMessage { return messages.NewLoginMessage() })
	registerClientMessage(10108, func() messaging.ClientMessage { return messages.NewKeepAliveMessage() })
	registerClientMessage(10212, func() messaging.ClientMessage { return messages.NewChangeAvatarNameMessage() })
	registerClientMessage(10107, func() messaging.ClientMessage { return messages.NewClientCapabilitiesMessage() })
	registerClientMessage(14102, func() messaging.ClientMessage { return messages.NewEndClientTurnMessage() })
	registerClientMessage(14113, func() messaging.ClientMessage { return messages.NewAskProfileMessage() })
	registerClientMessage(14109, func() messaging.ClientMessage { return messages.NewGoHomeFromOfflineMessage() })
	registerClientMessage(14110, func() messaging.ClientMessage { return messages.NewAskForBattleEndMessage() })
	registerClientMessage(14303, func() messaging.ClientMessage { return messages.NewAskForJoinableAlliancesMessage() })
	registerClientMessage(14302, func() messaging.ClientMessage { return messages.NewAskForAllianceDataMessage() })
	registerClientMessage(14315, func() messaging.ClientMessage { return messages.NewAllianceChatMessage() })
	registerClientMessage(14316, func() messaging.ClientMessage { return messages.NewAllianceEditMessage() })
	registerClientMessage(14305, func() messaging.ClientMessage { return messages.NewAllianceJoinMessage() })
	registerClientMessage(14308, func() messaging.ClientMessage { return messages.NewAllianceLeaveMessage() })
	registerClientMessage(14301, func() messaging.ClientMessage { return messages.NewAllianceCreateMessage() })
	registerClientMessage(14403, func() messaging.ClientMessage { return messages.NewGetLeaderboardMessage() })
	registerClientMessage(14350, func() messaging.ClientMessage { return messages.NewTeamCreateMessage() })
	registerClientMessage(14354, func() messaging.ClientMessage { return messages.NewTeamChangeMemberSettingsMessage() })
	registerClientMessage(14355, func() messaging.ClientMessage { return messages.NewTeamSetReadyMessage() })
	registerClientMessage(14361, func() messaging.ClientMessage { return messages.NewTeamMemberStatusMessage() })
	registerClientMessage(14352, func() messaging.ClientMessage { return messages.NewTeamKickMessage() })
	registerClientMessage(14353, func() messaging.ClientMessage { return messages.NewTeamMemberLeaveMessage() })
	registerClientMessage(14358, func() messaging.ClientMessage { return messages.NewTeamJoinMessage() })
	registerClientMessage(14356, func() messaging.ClientMessage { return messages.NewTeamTogglePracticeMessage() })
	registerClientMessage(14357, func() messaging.ClientMessage { return messages.NewTeamToggleMemberSideMessage() })
	registerClientMessage(14359, func() messaging.ClientMessage { return messages.NewTeamChatMessage() })
	registerClientMessage(14360, func() messaging.ClientMessage { return messages.NewTeamPostAdMessage() })
	registerClientMessage(14306, func() messaging.ClientMessage { return messages.NewAlliancePromoteMessage() })
	registerClientMessage(14307, func() messaging.ClientMessage { return messages.NewAllianceKickMessage() })
}
