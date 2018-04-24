# Smarter tado° AC

Not long ago, [tado°](https://www.tado.com/) added a new way to control AC units, one that finally took into account the temperature as measured by its smart AC thermostat.

I was excited about this, but unfortunately this mode comes with a big shortcoming that led me to skip the upgrade entirely.

This repository contains an alternative way for tado° AC to take into account temperature readings by the smart AC thermostat to help save energy.

Hopefully this inspires tado° to further improve its AC product, or open up an API flexible enough that it allows others to better implement alternative control modes.


## How does tado° AC work

tado° AC can work in one of two modes, the original “non-thermostatic” mode and the newer “thermostatic” mode.

### Non-thermostatic control

In the original, “non-thermostatic” mode, tado° AC works like most third-party “universal” remote controls. You configure the brand and model of your AC remote, and then you’re able to control most functions of your AC unit remotely through the tado° app. Some features of your AC might be absent.

Using tado° AC to control your unit adds the ability to setup a weekly schedule (which is much nicer to do through a mobile app), and helps save energy by automatically turning your AC off when everyone leaves home.

### Thermostatic control

With the newer, default, “thermostatic” mode, you manually record your preferred setting for “cooling” and “heating”. You’re advised to set the lowest possible temperature for “cooling” (and highest possible for “heating”). You can choose your prefered fan speed, vane position, etc. There’s only four commands to record, and almost all ACs are supported.

Then, tado° uses its own thermostat to periodically decide whether to turn your AC on/off. You can setup a weekly schedule, and tado° turns your AC off when no one is home.

You can opt-out from “thermostatic” mode by contacting support. But why would you?

#### Downsides of thermostatic control

The major downside of “thermostatic” mode is that it works against modern inverter ACs.

Inverter type ACs are able to vary the speed of the compressor to deliver just the right amount of cooling/heating power needed to maintain your preferred temperature. This has several advantages, such as improved confort from reduced temperature fluctuations, but also much improved efficiency.

But in “thermostatic” mode, you’re encouraged to set your AC to something like 16°C (that’s the command you record on setup). Then, if you set cooling to a more reasonable 24°C in the app, tado° will turn your AC on/off every 10min if temperature goes above/bellow 24°C. While it is on, your AC is working as hard as it can to needlessly reach those 16°C.

The other disadvantage is that, every time tado° changes its mind, your AC beeps (which can be annoying in a bedroom, at night).

## How does the code in this repository work

Here you will find a [Go](https://golang.org/) app that should be executed at least every 2-5min. It should be trivial to build for anyone familiar with Go, and I can gladly provide binaries for any platforms supported by the Go compiler on request.

You’re supposed configure your tado° username/password in the [config.json](config.json) file, and setup a “scheduled task” (Windows) or a “cron” job (Unix) to run it regularly.

Then, every 2-5min, this logs into tado° using the “beta” API partially documented [here](http://blog.scphillips.com/posts/2017/01/the-tado-api-v2/), reads status and settings for each of your AC zones, and decides if any action should be taken.

There are only two actions that this application might take for a zone:
1. turn off your AC, if it finds it is already too cold/hot for it to run. 
2. boost your fan speed, if it is set to auto and temperature is too hot/cold.

It never messes with the temperature or any other settings of your AC.

### Working scenarios

Imagine this spring/summer you setup your AC through tado° for cooling, with auto fan speed at 24°C. You are not using “thermostatic” mode.

So, today, it’s cloudy outside, and the temperature isn’t that high inside: it’s actually 24°C this morning. This would be a day you would never, yourself, turn your AC on. But because it’s in the schedule, “non-thermostatic” mode will turn on your AC. Your AC won’t do much, but it’s annoying that it turns on at all. This app prevents this from happening: if the measured temperature is close enough to your desired setting, the AC is preemptively kept off.

Further into summer, when it’s actually hot, your AC turns on, and cools the room down. But then the sun sets, and it’s now much cooler outside. So the AC eventually overshoots your set temperature: it’s now 23°C inside and it’s not getting any hotter. This app will turn off your AC for at least 15min if your set temperature is overshot.

But let’s say you have a big room, and it’s getting really hot. Near the AC unit it’s cool enough, but in the middle of the room where your thermostat is placed it’s still hot (more than 2°C over your set temperature). This app will manually boost fan speed to compensate.

### Conclusion

All these actions are meant to save power and actually improve efficiency, something “thermostatic” mode IMHO fails to do for more modern inverter type ACs. It would, obviously, be much better if tado° implemented something closer to this themselves. Again, hopefully, this inspires them to improve.

All that said, this has been working successfully in my own home throughout this fall/winter, and early spring. I have 3 ACs and thermostats in 3 different rooms. Now that it’s getting hotter, the ACs hardly ever turn on (heating mode). I finally reached the goal of a “set it and forget it” that works sensibly 90% of the time. I now expect to have to change settings only twice a year (for the hot/cold seasons).
