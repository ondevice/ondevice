#!/bin/bash -e

ondevice help
ondevice help ssh

echo 5h42l5xylznw | ondevice login --batch=demo

ondevice list
ondevice list --json

# TODO add some more tests
