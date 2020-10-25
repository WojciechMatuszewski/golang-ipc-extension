# Golang IPC extension

This is an example of setting up an lambda extension with golang. Lambda can hit the endpoint of a server which is created by the extension to retrieve SSM parameter.
Subsequent requests are cached. This is great since it works for every runtime!
