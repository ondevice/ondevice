ondevice.io command line client
===============================

This is the official [ondevice.io][ondevice.io] commandline client, written in [Go][go].

ondevice.io is a service that simplifies access to your SSH-enabled devices.


## Getting started:

The easiest way to install the ondevice client is by using the following one-liner:

```sh
curl -sSL https://get.ondevice.io/ | sudo bash -
``` 






It replaces the original [python client][pyClient]. The main reason for the rewrite was
that go applications can be shipped as single binary, dramatically simplifying my
packaging job :)  
Being a modern, statically typed and compiled language helps Go's case too
(but I might be biased in that regard)

I've just started writing this client, so it's far from being on par in with the python
client in terms of features. The client side of the tunnel is mostly done though.

[ondevice.io]: https://ondevice.io/
[go]: https://golang.org
[pyClient]: https://github.com/ondevice/ondevice-client/
