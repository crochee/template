package quota

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"template/pkg/logger"
	"template/pkg/redis"
	"template/pkg/resp"
)

func RegisterAPI(router *gin.Engine) {
	router.GET("/used_quota", UserUsedQuota)
	router.POST("/used_quota", RefreshUserUsedQuota)
}

type UsedQuotaRes struct {
	ResourceName     string // 资源名称
	RealQuota        int    // 真实配额数量
	RealUsed         int    // 真实使用量数据
	CacheQuota       int    // 缓存中的配额数据
	CacheUsed        int    // 缓存中的使用数据
	RefreshCacheTime int64  // 缓存上一次刷新时间
	IsError          bool   // 数据是否有异常
	ErrorMsg         string // 错误信息
	HasLock          string // 账号是否 存在锁，读锁表示正在预占数据，写锁表示正在刷新数据
}

func UserUsedQuota(c *gin.Context) {
	ctx := c.Request.Context()
	accounts := c.QueryArray("account")
	isFilter := c.Query("filter_error")
	if len(accounts) == 0 {
		client := redis.NewRedisClient()
		_, err := client.ZRemRangeByScore(QuotaAccountZSetKey, "0",
			strconv.FormatInt(time.Now().Add(time.Hour*-24).UnixMilli(), 10))
		if err != nil {
			logger.FromContext(ctx).Err(err).Msg("clear redis expire account fail")
			resp.Error(c, err)
			return
		}
		accounts, err = client.ZRange(QuotaAccountZSetKey, 0, -1)
		if err != nil {
			logger.FromContext(ctx).Err(err).Msg("get redis all account fail")
			resp.Error(c, err)
			return
		}
	}
	result := make(map[string][]UsedQuotaRes, len(accounts))
	for i := len(accounts) - 1; i >= 0; i-- {
		account := accounts[i]
		var res []UsedQuotaRes
		RangeRefreshHandler(func(resource string, handler resourceHandler) {
			var errMsg string
			used, err := handler.QueryUsed(ctx, account)
			if err != nil {
				errMsg = "used err: " + err.Error() + ", "
			}
			quota, err := handler.QueryQuota(ctx, account)
			if err != nil {
				errMsg = "quota err: " + err.Error() + ", "
			}
			usedCache, quotaCache, updateTime, errCache := GetAccountCacheResourceUsedQuota(ctx, account, resource)
			if errCache != nil {
				errMsg = errMsg + "cache err: " + errCache.Error()
			}
			if isFilter == "1" && errMsg == "" &&
				used == usedCache && quota == quotaCache {
				return
			}
			lockModel, _ := ExitLock(ctx, fmt.Sprintf(LockKey, account))
			res = append(res, UsedQuotaRes{
				ResourceName:     resource,
				RealQuota:        quota,
				RealUsed:         used,
				CacheQuota:       quotaCache,
				CacheUsed:        usedCache,
				RefreshCacheTime: updateTime,
				IsError:          errMsg != "",
				ErrorMsg:         errMsg,
				HasLock:          lockModel,
			})
		})
		if isFilter == "1" && len(res) == 0 {
			continue
		}
		result[account] = res
	}
	c.JSON(http.StatusOK, &result)
}

func RefreshUserUsedQuota(c *gin.Context) {
	ctx := c.Request.Context()
	accounts := c.QueryArray("account")
	resources := make([]string, 0, len(handlerMap))
	for resource := range handlerMap {
		resources = append(resources, resource)
	}
	for _, act := range accounts {
		err := RefreshAccountUsedQuota(ctx, act, true, resources...)
		if err != nil {
			logger.FromContext(ctx).Err(err).Str("account", act).Strs("resources", resources).Msg("account init refresh account used quota fail")
			resp.Error(c, errors.New("account "+act+" refresh used quota fail,err:"+err.Error()))
			return
		}
	}
	c.Status(http.StatusNoContent)
}
