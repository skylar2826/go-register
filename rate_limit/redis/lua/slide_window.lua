-- 获取当前服务是否存在
local key = KEYS[1]
-- 当前时间
local now = tonumber(ARGV[1])
-- 窗口大小
local interval = tonumber(ARGV[2])
-- 传输速率
local rate = tonumber(ARGV[3])
-- 起始位置
local min = now - interval

-- 删除起止位置前的数据
redis.call('ZREMRANGEBYSCORE', key,  '-inf', min)
-- 计算当前速率
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

if cnt < rate then
    redis.call("ZADD", key, now, now)
    redis.call("EXPIRE", key, interval)
    return "false"
else
    return "true"
end