package storage

import (
	"context"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Caller struct {
	// 上层调用者资源ID
	CallerID string `json:"caller_id,omitempty" gorm:"type:varchar(55);comment:上层调用者资源ID;index"`
	// 上层调用者资源编码
	CallerCode string `json:"caller_code,omitempty" gorm:"type:varchar(55);comment:上层调用者资源编码;index"`
	// 上层用户ID
	CallerUser string `json:"caller_user,omitempty" gorm:"type:varchar(55);comment;上层用户ID;index"`
	// 上层调用者资源扩展参数
	CallerExtra datatypes.JSON `json:"caller_extra,omitempty" gorm:"type:json;comment:上层调用者扩展参数"`
}

var (
	callerGetCallerID    func(ctx context.Context) string
	callerGetcallerCode  func(ctx context.Context) string
	callerGetcallerUser  func(ctx context.Context) string
	callerGetCallerExtra func(ctx context.Context) string
)

func CallerSetFunc(getCallerID, getcallerCode, getcallerUser, getCallerExtra func(ctx context.Context) string) {
	callerGetCallerID = getCallerID
	callerGetcallerCode = getcallerCode
	callerGetcallerUser = getcallerUser
	callerGetCallerExtra = getCallerExtra
}

// FillCaller fill the caller object with the value read from global context.
func (r *Caller) FillCaller(ctx context.Context) {
	if callerGetCallerID != nil {
		r.CallerID = callerGetCallerID(ctx)
	}
	if callerGetcallerCode != nil {
		r.CallerCode = callerGetcallerCode(ctx)
	}
	if callerGetcallerUser != nil {
		r.CallerUser = callerGetcallerUser(ctx)
	}
	if callerGetCallerExtra != nil {
		if extra := callerGetCallerExtra(ctx); extra != "" {
			r.CallerExtra = datatypes.JSON(extra)
		}
	}
}

func (r *Caller) BeforeCreate(db *gorm.DB) error {
	r.FillCaller(db.Statement.Context)
	return nil
}
