<p align="center">
	<a href="https://shopware.com.com"><img src="https://assets.shopware.com/media/logos/shopware_logo_blue.svg" alt="Shopware" width="450"></a>
</p>
<hr>
<h3 align="center">Foundation for Shopware apps based on Go</h3>
<p align="center">This server library provides and easy to set-up web sever for apps with a Go backend server.</p>


### Menu

- [Features](#features)
- [Quick start](#quick-start)
- [About](#about)


## Features

- **Automated handshake** for easy installation in Shopware
- **Easy configuration** with no additional router set-up needed
- **Generic endpoints** for admin action buttons and webhooks 
- Written in Go, a language with high memory safety guarantees 

## Quick start

### Installing

To start using the app server, install Go and run `go get`:

```sh
$ go get github.com/shopwareLabs/GoAppserver
```

### Storage engines

The app server comes with two storage engines included, in-memory and [bbolt](https://github.com/etcd-io/bbolt).

#### in-memory (default)

This storages resets on every restart of the server and should only be used for quick-start purposes. 
All information are lost when the process is killed. This storage is used by default.

#### bbolt

bbolt is a key/value store based on files provide a simple, fast, and reliable database for projects 
that don't require a full database server such as Postgres or MySQL

```go
store, err := appserver.NewBBoltStore("./mydb.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

srv := appserver.NewServer(
    "https://appserver.com",
    "AppName",
    "AppSecret",
    appserver.WithCredentialStore(store),
)

log.Fatal(srv.Start(":10100"))
```

### Events

To listen on an event, add the event to your `manifest.xml` file in your app and point it to `/webhook`. 
Verifying the signature is done automatically done for you.

**manifest.xml**

```xml
<webhooks>
    <webhook name="orderCompleted" url="https://appserver.com/webhook" event="checkout.order.placed"/>
</webhooks>
```

You can then register an event listener to the appserver:

```go
srv.Event("checkout.order.placed", func(webhook appserver.WebhookRequest, api *appserver.ApiClient) error {
    // do something on this event

    return nil
})
``` 

### Action buttons

To listen on a click on an action button, add the action button to your `manifest.xml` file in your app and point it to `/action`. 
Verifying the signature is done automatically done for you.

**manifest.xml**

```xml
<admin>
    <action-button action="doSomething" entity="product" view="detail" url="https://appserver.com/action">
        <label>do something</label>
    </action-button>
</admin>
```

You can then register an admin action listener to the appserver:

```go
srv.Action("product", "doSomething", func(action appserver.ActionRequest, api *appserver.ApiClient) error {
    // do something when someone clicks the action button

    return nil
})
``` 

### Full example

Here is a full example on an appserver that listens to events and action buttons.

```go
package main

import (
	appserver "github.com/shopwareLabs/GoAppserver"
	"log"
)

func main() {
	srv := appserver.NewServer(
		"https://appserver.com",
		"AppName",
		"AppSecret",
	)

	// event listener
	srv.Event("checkout.order.placed", func(webhook appserver.WebhookRequest, api *appserver.ApiClient) error {
		// do something on this event
		
		return nil
	})

	// action buttons
	srv.Action("product", "doSomething", func(action appserver.ActionRequest, api *appserver.ApiClient) error {
		// do something when someone clicks the action button
		
		return nil
	})

	log.Fatal(srv.Start(":10100"))
}
```

## About

GoAppserver is a project of [shopware AG](https://shopware.com).
