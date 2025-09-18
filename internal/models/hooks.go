package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BeforeCreate hooks for all models that have UUID primary keys

func (p *Platform) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (gv *GameVersion) BeforeCreate(tx *gorm.DB) error {
	if gv.ID == uuid.Nil {
		gv.ID = uuid.New()
	}
	return nil
}

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (rm *RoomMember) BeforeCreate(tx *gorm.DB) error {
	if rm.ID == uuid.Nil {
		rm.ID = uuid.New()
	}
	return nil
}

func (rm *RoomMessage) BeforeCreate(tx *gorm.DB) error {
	if rm.ID == uuid.Nil {
		rm.ID = uuid.New()
	}
	return nil
}

func (mr *MessageReaction) BeforeCreate(tx *gorm.DB) error {
	if mr.ID == uuid.Nil {
		mr.ID = uuid.New()
	}
	return nil
}

func (rt *ReactionType) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

func (ub *UserBlock) BeforeCreate(tx *gorm.DB) error {
	if ub.ID == uuid.Nil {
		ub.ID = uuid.New()
	}
	return nil
}

func (pn *PlayerName) BeforeCreate(tx *gorm.DB) error {
	if pn.ID == uuid.Nil {
		pn.ID = uuid.New()
	}
	return nil
}

func (uf *UserFollow) BeforeCreate(tx *gorm.DB) error {
	if uf.ID == uuid.Nil {
		uf.ID = uuid.New()
	}
	return nil
}

func (ua *UserActivity) BeforeCreate(tx *gorm.DB) error {
	if ua.ID == uuid.Nil {
		ua.ID = uuid.New()
	}
	return nil
}

func (rl *RoomLog) BeforeCreate(tx *gorm.DB) error {
	if rl.ID == uuid.Nil {
		rl.ID = uuid.New()
	}
	return nil
}

func (pr *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}
