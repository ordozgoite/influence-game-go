package rooms

import "errors"

type CreateRoomDTO struct {
	Nickname string `json:"nickname"`
}

func (dto *CreateRoomDTO) Validate() error {
	if dto.Nickname == "" {
		return errors.New("nickname_is_required")
	}
	return nil
}

type JoinRoomDTO struct {
	Nickname string `json:"nickname"`
}

func (dto *JoinRoomDTO) Validate() error {
	if dto.Nickname == "" {
		return errors.New("nickname_is_required")
	}
	return nil
}
