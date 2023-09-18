package dgc

import (
	"errors"

	dg "github.com/bwmarrin/discordgo"
)

type Client struct {
	DG *dg.Session
}

// Member returns a guild's member based on the specific guild and user IDs.
// Tries to use the local cache, if that fails, makes an API call.
func (c *Client) Member(guildID, userID string) (*dg.Member, error) {
	if mem, err := c.DG.State.Member(guildID, userID); err == nil {
		return mem, nil
	}
	return c.DG.GuildMember(guildID, userID)
}

// Channel returns a channel based on the specific channel ID.
// Tries to use the local cache, if that fails, makes an API call.
func (c *Client) Channel(channelID string) (*dg.Channel, error) {
	if ch, err := c.DG.State.Channel(channelID); err == nil {
		return ch, nil
	}
	return c.DG.Channel(channelID)
}

// Guild returns a guild based on the specific guild ID.
// Tries to use the local cache, if that fails, makes an API call.
func (c *Client) Guild(guildID string) (*dg.Guild, error) {
	if g, err := c.DG.State.Guild(guildID); err == nil {
		return g, nil
	}
	return c.DG.Guild(guildID)
}

// Role returns a role based on the specific guild and role IDs.
// Tries to use the local cache, if that fails, makes an API call.
func (c *Client) Role(guildID, roleID string) (*dg.Role, error) {
	if r, err := c.DG.State.Role(guildID, roleID); err == nil {
		return r, nil
	}

	rs, err := c.DG.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		if r.ID == roleID {
			return r, nil
		}
	}
	return nil, errors.New("couldn't find role")
}

// MemberAllowed returns true iff:
// - guildID is empty (aka a DM)
// - all the user's roles combined have all permissions in the perms bitfield
// - any of the user's roles has the administrator permission
// - the user is the owner of the guild
func (c *Client) MemberAllowed(guildID, userID string, perms int64) (bool, error) {
	if guildID == "" {
		return true, nil
	}

	mem, err := c.Member(guildID, userID)
	if err != nil {
		return false, err
	}

	// Both check if the individual role has all permissions (for efficiency,
	// as it is not uncommon for users to have a lot of roles), and the combined
	// roles' permissions
	var combined int64
	for _, roleID := range mem.Roles {
		r, err := c.Role(guildID, roleID)
		if err != nil {
			return false, err
		}
		if r.Permissions&dg.PermissionAdministrator != 0 {
			return true, nil
		}
		if r.Permissions&perms != 0 {
			return true, nil
		}
		combined |= r.Permissions
	}
	// if perms is contained in combined
	if perms&combined == perms {
		return true, nil
	}

	g, err := c.Guild(guildID)
	if err != nil {
		return false, err
	}
	return g.OwnerID == userID, nil
}

// VoiceState returns a voice state by guild and user ID.
// Tries to use the local cache, if that fails, makes an API call.
func (c *Client) VoiceState(guildID, userID string) (*dg.VoiceState, error) {
	if vs, err := c.DG.State.VoiceState(guildID, userID); err == nil {
		return vs, nil
	}

	g, err := c.Guild(guildID)
	if err != nil {
		return nil, err
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == userID {
			return vs, nil
		}
	}
	return nil, errors.New("could not find user's voice state")
}

// VoiceJoin joins the same voice channel in guild as user.
func (c *Client) VoiceJoin(guildID, userID string) (*dg.VoiceConnection, error) {
	vs, err := c.VoiceState(guildID, userID)
	if err != nil {
		return nil, err
	}
	return c.DG.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
}
