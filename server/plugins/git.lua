-- defines a factorial function
function git (args)
    print("translating...")
    return string.format( "git clone -b %s git@github.com:%s ", args['branch'], args['repo'] )
end