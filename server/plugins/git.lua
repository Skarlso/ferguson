-- defines a factorial function
function git (repo, branch)
    print("translating...")
    return string.format( "git clone -b %s git@github.com:%s ", repo, branch )
end