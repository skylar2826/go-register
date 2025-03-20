-- 检查当前服务是否存在
local val = redis.call('get', KEYS[1])
-- 获取窗口大小
local expiration = tonumber(ARGV[1])
-- 获取最大传输速率
local rate = tonumber(ARGV[2])

if val == false then
    if rate < 1 then
        return "true"
    else
        redis.call('set', KEYS[1], 1, 'PX', expiration)
        return "false"
    end
elseif tonumber(val) < rate then
    redis.call('incr', KEYS[1])
    return "false"
else
    return "true"
end
