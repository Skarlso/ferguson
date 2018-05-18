-- defines a factorial function
function func (n)
    if n == 0 then
        return 1
    else
        return n * func(n-1)
    end
    end
