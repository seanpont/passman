=======
passman
=======

A Password Manager that happens to be written in Go.

Passwords are encrypted using Blowfish, then base-64'd to generate a handy little text file that you can save anywhere. Save it to a dropbox folder to sync passwords between computers. Email a copy to yourself as a backup. 

Usage: passman [options] <command> [<args>]

Available commands:
   add         Add a service
   config      Configure the password manager
   cp          Copy the password for a service to the clipboard
   i           Interactive mode (good for adding lots of passwords)
   init        Initialize the password manager
   ls          List services
   repass      Change the encryption password
   rm          Remove a service
   show        Show password for a service
   
------------
Installation
------------

- install go (http://golang.org/)
- go get github.com/seanpont/passman
- cd src/github.com/seanpont/passman
- go install .
    * this will install an executable named passman in $GOPATH/bin
- passman init

Enjoy!

