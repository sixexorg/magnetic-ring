local a = 1;
b = 2;
if a > b then
    print("-------------")
else
    print("=============")
end
function max()
    print("max func be called")
    return 6
end


local _M = {}
_M._VERSION = '1.0'

local mt = {__index = _M}



function init()
    local a = 1;
    local b = 2;
    max()
    if a > b then
        print("-------------")
    else
        print("=============")
    end
    mytable = {}
    mytable["key"] = "key"
    mytable[2] = 6
--    store("a",a)
    store("mt",mt)
    store("mytalbe",mytable)
end

function _M.Withdraw( from, to, value )
    print(string.format("Withdraw from %s to %s, the amount is %d", from, to, value))
end

function _M.Transfer( from, to, value )
    print(string.format("Account %s initiates a transfer to account %s, the amount is %d", from, to, value))
end

function _M.Approval( from, to, amount )
    print(string.format("Account %s approves another account %s to allow it to be used for %d", from, to, amount))
end
init()
return _M

