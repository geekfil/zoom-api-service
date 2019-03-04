#!/bin/bash
go vet && git add . && git commit -am update && git push
