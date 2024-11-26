---
title: Client Libraries

layout: default
nav_order: 6
---
# Client Libraries

Since `propeller` is built with `protobuf`, it's easy to generate client libraries in different languages. 

List of currently supported languages.

## Golang

```
import "github.com/CRED-CLUB/propeller/tree/main/rpc/push/v1"
```

## Java

Java artifacts are published on maven central.

Link : [https://central.sonatype.com/artifact/io.github.abhishekvrshny/propeller](https://central.sonatype.com/artifact/io.github.abhishekvrshny/propeller)

### Apache Maven

```
<dependency>
    <groupId>io.github.abhishekvrshny</groupId>
    <artifactId>propeller</artifactId>
    <version>0.0.1</version>
</dependency>
```

### Gradle

```
implementation group: 'io.github.abhishekvrshny', name: 'propeller', version: '0.0.1'

```

{: .note }
For support for any other language, please raise a [github issue](https://github.com/CRED-CLUB/propeller/issues)
---
