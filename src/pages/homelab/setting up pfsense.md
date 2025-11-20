---
layout: /src/layouts/BaseLayout.astro
author: avolent
title: Setting up pfSense
date: May 2024
publish: 'true'
---

## Summary

The following page goes over how I configure my pfSense at home and what steps it takes to get there.

## Assumptions

- You are using pfSense
- Basic understanding of networking & pfSense

## Contents

1. [Useful Links](#useful-links)
2. [Initial Setup](#initial-setup)
3. [UPNP](#upnp)
4. [Port Forwarding](#port-forwarding)
5. [References](#references)

## Useful Links

- [pfsense Web Page](https://192.168.1.1/)
- [Buffer Bloat Setup Video](https://www.youtube.com/watch?v=iXqExAALzR8&ab_channel=LawrenceSystems)
- [Pfsense pfblockerng Setup](https://www.youtube.com/watch?v=xizAeAqYde4&ab_channel=LawrenceSystems)
- [Wireguard Setup](https://www.youtube.com/watch?v=8jQ5UE_7xds&ab_channel=LawrenceSystems)

## Default Pfsense Login

**Username:** admin  
**Password:** pfsense

## Initial Setup

Useful video for setup of pfsense and basic settings. [^3]

### 1. [/System/Advanced/Admin Access](https://192.168.1.1/system_advanced_admin.php)

Enable **Display page name first in browser tab**.  
Enable **Secure Shell** if required.

### 2. [/Advanced/Firewall & NAT](https://192.168.1.1/system_advanced_admin.php)

Update **Firewall Maximum Table Entries** to 10,000,000.  
Enable **IP Random id generation**.

### 3. [/Advanced/Misc](https://192.168.1.1/system_advanced_misc.php)

Enable **PowerD**.  
Enable **Cryptographic Hardware**.  
Enable **Thermal Sensors**.  
Enable **Do NOT send Netgate Device ID with user agent**.

### 4. [/System/General Setup](https://192.168.1.1/system.php)

**Theme** = pfsense Dark.  
**Top Navigation** = Fixed.  
**Dashboard Columns** = 3.

### 5. [/System/Package Manager/Available Packages](https://192.168.1.1/pkg_mgr.php)

```
darkstat (Optional)
iperf (Optional)
nmap (Optional)
pfblockerng-devel (Important)
```

### 6. [/System/User Manager/Users](https://192.168.1.1/system_usermanager.php)

Add a new local user and attach them to the administrator group.  
Log out and log into the new user.

### 8. [/Interfaces/WAN](https://192.168.1.1/interfaces.php?if=wan)

Changing port assignments can be done in [/Interfaces/Assignments](https://192.168.1.1/interfaces_assign.php).  
Most things are fine as default in here.  
If you need/want IPV6 change **DHCPv6 Prefix Delegation size** to the size your internet provider has given you.

### 7. [/Services/DHCP Server/LAN](https://192.168.1.1/services_dhcp.php)

Enable **Change DHCP display lease time from UTC to local time** (Do the same for IPv6).  
Assign your **Static IP** addresses.

### 8. [/Services/DNS Resolver/General Settings](https://192.168.1.1/services_unbound.php)

Enable **DNSSEC** Support.  
Enable **Python Module** (Leave the settings default) _If you want regex blocking_.

### 9. [/Services/Dynamic DNS/Dynamic DNS Clients](https://192.168.1.1/services_dyndns.php)

Create a Cloudflare ddns client.

### 10. [/Firewall/pfBlockerNG](https://192.168.1.1/pfblockerng/pfblockerng_general.php) [^4]

**CRON Settings** = Once a day.

### 11. [/Firewall/pfBlockerNG/IP](https://192.168.1.1/pfblockerng/pfblockerng_ip.php) - Make Sure You Force Reload at the End to save Changes

Enable **Floating Rules**.  
Enable **Kill States**.  
Add **MaxMind License Key**.

### 12. [/Firewall/pfBlockerNG/DNSBL](https://192.168.1.1/pfblockerng/pfblockerng_dnsbl.php)

**DNSBL Mode** = Unbound Python Mode _If you want regex blocking_.  
**Regex Blocking** = Enable.  
**DNSBL IPs - List Action** = Deny Both.  
Add all your feeds and enable them.

MAKE SURE YOU **UPDATE/RELOAD ALL** TO MAKE SURE CHANGES ARE ACTIVE!!

If you get an error about running out of memory, this could be related to your firewall maximum table entry size. Since pfblockerNG is using a lot of rules you may need to increase even more as seen in a previous step. [^6]

### PFBlockerNG Feeds

The following feeds are what I currently use!

```bash
PRI1
PRI2
PRI1_6
Easylist
ADs
AD_Basic
Cryptojackers
Firebog_Advertising
```

## UPNP

Configuring the following is great for devices which have strict NAT type when playing games.  
Device used in this example is a **Nintendo Switch** [^1].

### 1. [/Services/DHCP Server/LAN](https://192.168.1.1/services_dhcp.php)

Assign device a static IP.

### 2. [/Services/UPnP & NAT-PMP](https://192.168.1.1/pkg_edit.php?xml=miniupnpd.xml&id=0)

Enable the following:  
**UPnP & NAT-PMP**  
**Allow UPnP Port Mapping**  
**Allow NAT-PMP Port Mapping**  
**Log packets handled by UPnP & NAT-PMP rules.**  
**Deny access to UPnP & NAT-PMP by default.**

Within **ACL Entries** add the device configured with the static IP like so.  
`allow 53-65535 192.168.1.246/32 53-65535`

### 3. [/Firewall/NAT/Outbound](https://192.168.1.1/firewall_nat_out.php)

Configure your Outbound NAT mode to **Hybrid**  
Add the following rule.  
**Interface:** WAN  
**Address Family:** IPv4 + IPv6 (IPv4 only if IPv6 not enabled)  
**Protocol:** Any  
**Source:** Network | deviceip/32 | no port range  
**Destination:** Any  
**Translation Address:** Interface Address  
**Translation Port or Range:** Empty | Check the static box

Then add a **description** to this rule.

Your device should now have a better NAT Type when playing games.

## Port Forwarding

Port forward so that you can host games on your internal network and share with your friends externally.

### 1. [/Firewall/NAT/Port Forward](https://192.168.1.1/firewall_nat.php)

Add a rule for a port you would like to forward.  
**Interface:** WAN  
**Address Family:** IPv4 (Unless you need IPv6)  
**Protocol:** TCP / UDP or Both  
**Destination:** WAN  
**Destination Port Range:** Either a single port or a range. (This is the external port that will be connected too.)  
**Redirect IP:** The device IP hosting the server.  
**Redirect Target Port:** This is the port in which your server is using. (Multiple rules required for multiple ports.  
**Description:** Description for the rule.  
**Filter rule association:** Add associated filter rule.

Save the rule.

> These ports should now be open to internet. BE CAREFUL and open only what is needed. Block unnecessary global subnets etc via PFBlockerNG.

## Buffer Bloat

todo  
[^2]

## Wireguard

todo  
[^5]

## VLANs

todo

## References

[^1]: [How to get open NAT](https://www.amixa.com/blog/2020/04/02/how-to-get-open-nat-with-xbox-or-xbox-one-and-pfsense-firewall/)

[^2]: [Buffer Bloat by Lawrence Systems](https://www.youtube.com/watch?v=iXqExAALzR8&ab_channel=LawrenceSystems)

[^3]: [2020 Pfsense Setup by Lawrence Systems](https://www.youtube.com/watch?v=fsdm5uc_LsU&t=662s)

[^4]: [pfblockerng Setup by Lawrence Systems](https://www.youtube.com/watch?v=xizAeAqYde4&ab_channel=LawrenceSystems)

[^5]: [Wireguard Setup by Lawrence Systems](https://www.youtube.com/watch?v=CXFbEbzFEXw&feature=emb_title&ab_channel=LawrenceSystems)

[^6]: [Firewall Table limit Issue](https://forum.netgate.com/topic/129127/ruleerror-there-were-errors-loading-the-rules-tmp-rules-debug-18-cannot-alloc/25)
