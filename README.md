# HiveKit [![Circle CI](https://circleci.com/gh/njpatel/hivekit.svg?style=svg)](https://circleci.com/gh/njpatel/hivekit)

# HiveKit [![Circle CI](https://circleci.com/gh/njpatel/hivekit.svg?style=svg)](https://circleci.com/gh/njpatel/hivekit)

HiveKit is a [HomeKit](https://developer.apple.com/homekit/) bridge to make the [Hive Active Heating](https://hivehome.com) system available as HomeKit accessories. This allows you to use your Hive with your iPhone, Apple Watch, and Siri.

HiveKit is based on the excellent [HomeControl](https://github.com/brutella/hc) by [@brutella](https://twitter.com/brutella).

## Overview
HiveKit is a daemon that registers itself as a HomeKit *bridge* with three *accessories*:

 * **Heating** *thermostat*
 * **Heating Boost** *switch*
 * **Hot Water** *switch*
 
These accessories expose some of the functionality of your Hive device to HomeKit. Once a connection is set up, you can view the status of these accessories & manipulate them via any HomeKit-compatible application and/or Siri, on your iPhone/iPad/Apple Watch.

HiveKit connects to the [hivehome.com](https://hivehome.com) web service to read & manipulate the Hive components in your home. Every installation of the Hive should come with a login to the aformentioned service (usually and email + password), and the same login details are used for the Hive application on iOS/Android.

**Note:** The way HomeKit works means that it has a very strict world-view of home automation accessories, and therefore HiveKit does it's best to fit the Hive device and it's capabilities into this world-view. This means that, while the most important features are available, not all the features of the Hive device have been mapped as yet. As HomeKit & HiveKit mature, these will be made available.

## Features
 * View & set the current & target temperature on the Hive via any HomeKit-enabled app and/or Siri
 * Boost the heating (and manipulate the boost) via an app and/or Siri
 * Boost the hot water (and manipulate the boost) via an app and/or Siri
 
## HomeKit App v. HomeKit Bridge
The terminology and architecture around HomeKit can be confusing and therefore hard to figure out how things fit together. 

**HiveKit** is a *HomeKit Bridge* daemon, it's meant to be run on your computer and allowed to go about it's business. It's run from the command-line and has no GUI itself.

**HomeKit** is the database/register on your phone that comes as part of iOS 8.0 and above. Again, it's lacks any UI itself, and is rather a service that runs in the background waiting for an *HomeKit app* to register some accessories with it. Once accessories have been registered, HomeKit springs to life and makes them available to Siri and relays commands back to the HomeKit App. Without a HomeKit App, HomeKit itself is useless.

A **HomeKit App** is an app downloaded from the App Store that can hook devices/bridges that support HomeKit on your home network to the HomeKit database on your phone. Without such an app, neither HomeKit or HiveKit are of any use. Some example apps are:

 * [Home](https://itunes.apple.com/app/id995994352)
 * [Devices](https://itunes.apple.com/gb/app/devices/id966877433?mt=8)
 
HiveKit is developed and tested with [Home](https://itunes.apple.com/app/id995994352). The developer of Home also produced the [open-source library](https://github.com/brutella/hc) that HiveKit uses to present itself as a HomeKit bridge. Supporting Home inadvertedly supports HiveKit :)
 
 
## Getting Started
HiveKit is written in [Go](http://golang.org/doc/install) and you will need a working Go 1.5 installation and workspace to build and run HiveKit.

1. Clone HiveKit
        git clone github.com/njpatel/hivekit.git
        cd hivekit 
2. Install HiveKit dependencies
3. 
