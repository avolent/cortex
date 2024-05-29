---
layout: /src/layouts/BaseLayout.astro
author: avolent
title: Voxelab
date: May 2024
publish: "true"
---

## Contents

1. [Useful Links](#useful-links)
2. [Updating your Printer](#updating-your-printer)
	1. [Preparation](#preparation)
	1. [Flashing the motherboard](#flashing-the-motherboard)
	1. [Flashing the display](#flashing-the-display)
3. [References](#references)

## Useful Links

- [Custom Firmware](https://github.com/alexqzd/Marlin/releases) - Video Tutorial [^3]
- [Stock Firmware](https://www.voxelab3dp.com/download) - Video Tutorial [^2]

## Updating your Printer

### Preparation

Important Sticky Note to read before starting [here](https://www.reddit.com/r/VoxelabAquila/comments/lvlzf2/sticky_post_with_links_to_important_posts/) [^1]
#### Determine which firmware your printer requires. (G32 or N32)

Explanation [here](https://www.reddit.com/r/VoxelabAquila/comments/ph1a7u/explanation_on_aquila_chips_from_voxelab_team/)

- Simple method
	Confirm whether you printer has a sticker on it with the model of chip.
![chip.jpg](/images/devices/voxelab/chip.jpg)
If your printer does not have a sticker, it is most likely the G32 or N32

- Effort Required
	Open the bottom of your printer and confirm by looking at the model number on the chip.
	![chip_onboard.jpeg](/images/devices/voxelab/chip_onboard.jpeg)

#### Download custom firmware

- Download the .bin file for your device version from [Custom Firmware](https://github.com/alexqzd/Marlin/releases)

**Types of version**
	- **<u>Bed Leveling Type</u>**
	BLTouch: Auto Bed Leveling
	Default: No Auto Bed Leveling
	Manual Mesh: Manual Bed Leveling, configure your own mesh.
	UBL: Unified Bed Leveling
	- **<u>Mesh Size</u>**
	3x3: 3 points by 3 points mesh.
	4x4: 4 points by 4 points mesh.
	and so on..
	- **<u>Chip Model</u>**
	G32 or N32: Determined earlier

- Also download the source code .zip file aswell.

#### Format the SD card

Open up windows explorer and format your SD card with the following settings

- Capacity : Doesn't matter dependent on SD card size.
- File System : FAT32
- Allocation Unit Size : 4096 Bytes
- Volume Name : Whatever you want!
- Quick Format : Checked

![format_settings](/images/devices/voxelab/format_settings.png)

#### Preparing for Flashing

- Extract the source code zip file.
- Go into the path "/Display Firmware/Firmware Sets". The folder set you can now see are basically the theme for the display. Pick one you like, copy it to your desktop and then rename it to just "DWIN_SET" (All capitals).
- Now for the .bin file you downloaded earlier, Make a new folder and name it "firmware" (All lowercase) and move it inside.
- Move both folders "firmware" and "DWIN_SET" to your freshly formatted SD Card.
End result should look something like this

![sd_folders](/images/devices/voxelab/sd_folders.png)

### Flashing the motherboard

- Make sure printer is off.
- Place SD card into Machine
- Turn on the printer
- Watch the screen and look for a loading bar. It will give you a check mark once complete.
*The screen will look really messy after flashing the mobo, this is normal as we havent flashed the display yet.*

### Flashing the display

- Make sure the printer is off.
- Remove the screen and the screws from the back.
- Open up the screen and remove the ribbon cable.
- Place the SD card in the back of the screen.
- Plug the ribbon cable back in.
- Turn the printer back on and wait for the screen to flash blue and then orange.
- Once showing orange we can put everything back to normal and we now have the new firmware.

**Firmware Upgrade Complete!**

## References

[^1]: [Reddit Sticky with important info](https://www.reddit.com/r/VoxelabAquila/comments/lvlzf2/sticky_post_with_links_to_important_posts/)
[^2]: [Stock Upgrade Video by 3DprintSOS](https://www.youtube.com/watch?v=6afQUIR6Dmo)
[^3]: [3rd Party Upgrade Video by 3DprintSOS](https://www.youtube.com/watch?v=sQFsnIyJ5BM)
