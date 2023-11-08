package cache

import (
	"context"
	"errors"

	"template/pkg/code"
	"template/pkg/logger"
	"template/pkg/redis"
)

// 定义账号信息的Hash Key
const (
	DCSAccountIDToNameKey = "dcs:order:id2name"
	DCSAccountNameToIDKey = "dcs:order:name2id"
	DCSPaasAccounts       = "dcs:order:paas"
	DCSInternalAccounts   = "dcs:order:internal"
)

// 账号类型
const (
	AccountTypePaas     = "paas"
	AccountTypeInternal = "internal"
	AccountTypeNormal   = "normal"
)

// 列表查询参数键值
const (
	QueryAccountID   = "account_id"
	QueryAccountName = "account_name"
)

func GetAccountIDs(ctx context.Context, names []string) (map[string]string, error) {
	var result = map[string]string{}
	// 当names为空列表，HGET会报错，此处判断一下
	if len(names) == 0 {
		return result, nil
	}
	names = removeRepeat(names)
	ids, err := redis.NewRedisClient().HMGet(DCSAccountNameToIDKey, names...)
	if err != nil {
		logger.FromContext(ctx).Err(err).Strs("account_names", names).Msg("Failed to get account_id from redis")
		return nil, code.ErrCodeRedisCacheOption.WithResult(err.Error())
	}

	result = make(map[string]string, len(names))
	for i, id := range ids {
		if id != nil {
			result[names[i]] = id.(string)
		}
	}
	return result, nil
}

func GetAccountName(ctx context.Context, id string) string {
	names, _ := GetAccountNames(ctx, id)
	return names[id]
}

func GetAccountNames(ctx context.Context, ids ...string) (map[string]string, error) {
	var result = map[string]string{}
	// 当ids为空列表，HGET会报错，此处判断一下
	if len(ids) == 0 {
		return result, nil
	}
	ids = removeRepeat(ids)
	names, err := redis.NewRedisClient().HMGet(DCSAccountIDToNameKey, ids...)
	if err != nil {
		logger.FromContext(ctx).Err(err).Strs("account_ids", ids).Msg("Failed to get account_name from redis")
		// 在redis出错的情况下， 保证逻辑可以走下去
		names = make([]interface{}, len(ids))
	}

	// 如果从redis查出来的名称个数少于ids的个数，则需要则填充一些nil数据，为了计算出emptyIds
	if len(names) < len(ids) {
		lackNames := make([]interface{}, len(ids)-len(names))
		names = append(names, lackNames...)
	}

	// 统计emptyIds, 通过从redis查出来的names来计算哪些ids，还需要再从woden层的orders接口中获取
	var emptyIds []string
	result = make(map[string]string, len(ids))
	for i, name := range names {
		if name != nil {
			result[ids[i]] = name.(string)
			continue
		}
		emptyIds = append(emptyIds, ids[i])
	}

	// 如果emptyIds大于0， 则表示这些IDs需要再从woden层的orders接口获取数据
	if len(emptyIds) > 0 {
		accounts, err := accountCache.GetAccount(ctx, emptyIds...)
		if err != nil {
			logger.FromContext(ctx).Err(err).Strs("account ids", emptyIds).Msg("get account info fail")
		}
		for _, account := range accounts {
			if account.AccountID != "" {
				result[account.AccountID] = account.AccountName
			}
		}
		err = accountCache.RefreshAccount(ctx, accounts...)
		if err != nil {
			logger.FromContext(ctx).Err(err).Interface("account infos", accounts).Msg("refresh account fail")
		}
	}

	return result, nil
}

func removeRepeat(accounts []string) []string {
	accountMap := make(map[string]struct{}, len(accounts))
	for _, account := range accounts {
		accountMap[account] = struct{}{}
	}
	accounts = make([]string, 0, len(accountMap))
	for account := range accountMap {
		accounts = append(accounts, account)
	}
	return accounts
}

type AccountInfo struct {
	AccountID   string
	AccountName string
	AccountType string
}

var accountCache AccountCache

func RegisterAccountGetFunc(getAccount func(ctx context.Context, accounts ...string) ([]AccountInfo, error)) {
	accountCache.getAccount = getAccount
}

func RefreshAccount(ctx context.Context, accountInfos ...AccountInfo) error {
	return accountCache.RefreshAccount(ctx, accountInfos...)
}

type AccountCache struct {
	getAccount func(ctx context.Context, accounts ...string) ([]AccountInfo, error)
}

func (r *AccountCache) GetAccount(ctx context.Context, accounts ...string) ([]AccountInfo, error) {
	if r.getAccount == nil {
		return nil, errors.New("not init get account func")
	}
	// 去重
	accounts = removeRepeat(accounts)
	return r.getAccount(ctx, accounts...)
}

func (r *AccountCache) RefreshAccount(ctx context.Context, accountInfos ...AccountInfo) error {
	if len(accountInfos) == 0 {
		return nil
	}
	nameMap := make(map[string]interface{}, len(accountInfos))
	idMap := make(map[string]interface{}, len(accountInfos))
	paasAccounts := make([]interface{}, 0)
	internalAccounts := make([]interface{}, 0)

	for _, account := range accountInfos {
		if account.AccountID != "" {
			nameMap[account.AccountID] = account.AccountName
		}
		if account.AccountName != "" {
			idMap[account.AccountName] = account.AccountID
		}
		switch account.AccountType {
		case AccountTypePaas:
			paasAccounts = append(paasAccounts, account.AccountID)
		case AccountTypeInternal:
			internalAccounts = append(internalAccounts, account.AccountID)
		default:

		}
	}

	if err := r.refreshAccountID(ctx, idMap); err != nil {
		return err
	}

	if err := r.refreshAccountName(ctx, nameMap); err != nil {
		return err
	}

	if err := r.refreshPaasAccount(ctx, paasAccounts); err != nil {
		return err
	}

	if err := r.refreshInternalAccount(ctx, internalAccounts); err != nil {
		return err
	}
	return nil
}

func (r *AccountCache) refreshAccountID(ctx context.Context, ids map[string]interface{}) error {

	cli := redis.NewRedisClient()
	if _, err := cli.HSet(DCSAccountNameToIDKey, ids); err != nil {
		logger.FromContext(ctx).Err(err).Msg("failed to refresh account id cache to redis")
		return code.ErrCodeRedisCacheOption.WithResult(err.Error())
	}

	return nil
}

func (r *AccountCache) refreshAccountName(ctx context.Context, names map[string]interface{}) error {
	if len(names) == 0 {
		return nil
	}

	cli := redis.NewRedisClient()
	if _, err := cli.HSet(DCSAccountIDToNameKey, names); err != nil {
		logger.FromContext(ctx).Err(err).Msg("failed to refresh account name cache to redis")
		return code.ErrCodeRedisCacheOption.WithResult(err.Error())
	}

	return nil
}

func (r *AccountCache) refreshPaasAccount(ctx context.Context, accounts []interface{}) error {
	if len(accounts) == 0 {
		return nil
	}

	cli := redis.NewRedisClient()
	if _, err := cli.SAdd(DCSPaasAccounts, accounts...); err != nil {
		logger.FromContext(ctx).Err(err).Msg("failed to refresh paas account cache to redis")
		return code.ErrCodeRedisCacheOption.WithResult(err.Error())
	}

	return nil
}

func (r *AccountCache) refreshInternalAccount(ctx context.Context, accounts []interface{}) error {
	if len(accounts) == 0 {
		return nil
	}
	cli := redis.NewRedisClient()
	if _, err := cli.SAdd(DCSInternalAccounts, accounts...); err != nil {
		logger.FromContext(ctx).Err(err).Msg("failed to refresh Internal account cache to redis")
		return code.ErrCodeRedisCacheOption.WithResult(err.Error())
	}
	return nil
}

// IsPaas 是否是PaaS用户, PaaS用户不需要预占配额
func IsPaas(account string) (bool, error) {
	cli := redis.NewRedisClient()
	return cli.SIsMember(DCSPaasAccounts, account)
}

// IsInternal 是否是Internal用户, Internal用户不需要预占配额
func IsInternal(account string) (bool, error) {
	cli := redis.NewRedisClient()
	return cli.SIsMember(DCSInternalAccounts, account)
}
