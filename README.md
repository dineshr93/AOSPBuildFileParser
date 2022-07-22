# AOSPBuildFileParser

This is a parser for .bp and .mk file

```
import (
bkparser "AOSPBuildFileParser/blueprint/parser"
mkparser "AOSPBuildFileParser/androidmk/parser"
)
```

## To build

```
cd AOSPBuildFileParser
go build -o aospparse.exe
```

## Usage

```
aospparse.exe path/to/Android.bp keyName_in_Android.bp


for ex: aospparse.exe Android.bp deps
api_surface.go
export.go
metadata.go
import.go

```

The Core Parsers are taken from Soong build system in the AOSP and are available under Apache-2.0 License to reuse and modify.

This repo is for tracing the Suppliers code specific AOSP dependencies for OSS compliance activity
