export default {
  common: {
    setting: "Setting",
    about: "About",
    loggedAs: "Logged as <b>{username}</b>",
    v2rayCoreStatus: "Status of v2ray-core",
    checkRunning: "Checking",
    isRunning: "Running",
    notRunning: "Stopped",
    notLogin: "Login please",
    latest: "Latest",
    local: "Local",
    success: "SUCCESS",
    fail: "FAIL",
    message: "Message",
    none: "none",
    optional: "optional"
  },
  welcome: {
    title: "Welcome",
    docker: "v2rayA service is running in Docker. Version: {version}",
    default: "v2rayA service is running. Version: {version}",
    newVersion: "Detected new version: {version}",
    messages: [
      "There is no server.",
      "You can create a server or import a subscription. Vmess, SS and SSR are supported."
    ]
  },
  v2ray: {
    start: "Start",
    stop: "Stop"
  },
  server: {
    name: "Server Name",
    address: "Server Address",
    protocol: "Protocol",
    latency: "Latency"
  },
  subscription: {
    host: "Host",
    remarks: "Remarks",
    timeLastUpdate: "Datetime of Last Update",
    numberServers: "Number of Servers"
  },
  operations: {
    name: "Operations",
    update: "Update",
    modify: "Modify",
    share: "Share",
    view: "View",
    delete: "Delete",
    create: "Create",
    import: "Import",
    connect: "Connect",
    disconnect: "Disconnect",
    login: "Login",
    logout: "Logout",
    configure: "Configure",
    cancel: "Cancel",
    saveApply: "Save and Apply",
    confirm: "Confirm",
    confirm2: "Carefully confirmed",
    save: "Save",
    copyLink: "COPY LINK",
    helpManual: "Help & Manual",
    yes: "Yes",
    no: "No",
    switchSite: "Switch to alternate site"
  },
  register: {
    title: "Create an admin account first",
    messages: [
      "Remember your admin account which is importantly used to login.",
      "Account information is stored in local. We never send information to any server.",
      "Once password was forgot, you could delete the config file and restart v2rayA service to reset."
    ]
  },
  login: {
    title: "Login",
    username: "Username",
    password: "Password"
  },
  setting: {
    transparentProxy: "Transparent Proxy",
    pacMode: "Traffic Splitting Mode of Rule Port",
    preventDnsSpoofing: "Prevent DNS Spoofing",
    mux: "Multiplex",
    autoUpdateSub: "Automatically Update Subscriptions",
    autoUpdateGfwlist: "Automatically Update GFWList",
    preferModeWhenUpdate: "Mode when Upadate Subscriptions and GFWList",
    ipForwardOn: "Share in LAN",
    enhancedModeOn: "Enhanced Mode",
    concurrency: "Concurrency",
    options: {
      global: "Proxy All Traffic",
      direct: "Direct",
      pac: "Depend on Rule Port",
      whitelistCn: "Proxy except CN Sites",
      gfwlist: "Proxy Only GFWList",
      sameAsPacMode: "The Same as the Rule Port",
      customRouting: "Customized Routing",
      antiDnsHijack: "Prevent DNS Hijack Only",
      forwardDnsRequest: "Prevent DNS Spoofing: Forward DNS Request",
      doh: "Prevent DNS Spoofing: DoH(dns-over-https)",
      default: "Keep Default",
      on: "On",
      off: "Off",
      updateSubWhenStart: "Update Subscriptions When Service Starts",
      updateGfwlistWhenStart: "Update GFWList When Service Starts",
      dependTransparentMode: "Depend on Transparent Proxy",
      closed: "Off"
    },
    messages: {
      gfwlist:
        "Based on modified time of file which sometimes is after latest version online.",
      transparentProxy:
        "If transparent proxy on, no extra configure needed and all TCP traffic will pass through the v2rayA. Providing proxy service to other computers as the gateway should make option 'Share in LAN' on.",
      pacMode: `Here you can set the splitting traffic rule of rule port. By default, "Rule of Splitting Traffic" port is 20172 and HTTP protocol.`,
      preventDnsSpoofing:
        "If there is a problem with transparent proxy, try setting 'Prevent DNS Spoofing' as 'Off' or turn on 'Enhanced Mode' (v0.7.0.2+)." +
        "★Forward DNS Request: DNS requests will be forwarded by proxy server." +
        "★DoH(dns-over-https, v2ray-core: 4.22.0+): Stable and fast DoH services are suggested." +
        "★Enhanced Mode(v0.7.0.2+) will replace the method of forwarding dns-query using iptables with DnsPoison method to deal with contaminated domain names.",
      tcpFastOpen:
        "Simplify TCP handshake process to speed up connection establishment. Risk of emphasizing characteristics of packets exists. Support vmess only now.",
      mux:
        "Multiplexing TCP connections to reduce the number of handshake, but it will affect the use cases with high throughput, such as watching videos, downloading, and test speed. " +
        "Risk of emphasizing characteristics of packets exists. Support vmess only now.",
      confirmEgressPorts: `<p>You are setting up transparent proxy across LANs, confirm egress port whitelist.</p>
                          <p>Whitelist:</p>
                          <p>TCP: {tcpPorts}</p>
                          <p>UDP: {udpPorts}</p>`
    }
  },
  customAddressPort: {
    title: "Address and Ports",
    serviceAddress: "Address of Service",
    portSocks5: "Port of SOCKS5",
    portHttp: "Port of HTTP",
    portHttpWithPac: "Port of HTTP(with Rule)",
    messages: [
      "Service address default as 0.0.0.0:2017 can be changed by setting environment variable <code>V2RAYA_ADDRESS</code> and command argument<code>--address</code>.",
      "If you start v2raya docker container with port mapping instead of <code>--network host</code>, you can remapping ports in this way.",
      "We cannot judge port occupations in docker mode. Confirm it by yourself.",
      "Zero means to close this port."
    ]
  },
  customRouting: {
    title: "Customize Routing Rule",
    defaultRoutingRule: "Default Routing Rule",
    sameAsDefaultRule: "the same as default rule",
    appendRule: "Append Rule",
    direct: "Direct",
    proxy: "Proxy",
    block: "Block",
    rule: "Rule",
    domainFile: "Domain File",
    typeRule: "Type of Rule",
    messages: {
      0: "v2rayA will recognize all SiteDat file in <b>{V2RayLocationAsset}</b>",
      1: 'To make a SiteDat file by yourself: <a href="https://github.com/ToutyRater/V2Ray-SiteDAT">ToutyRater/V2Ray-SiteDAT</a>',
      2: "Multi-select is supported.",
      noSiteDatFileFound: "No siteDat file found in {V2RayLocationAsset}",
      emptyRuleNotPermitted: "Empty rule is not permitted"
    }
  },
  doh: {
    title: "Configure DoH Server",
    dohPriorityList: "Priority list of DoH Servers",
    messages: [
      "DoH (DNS over HTTPS) can effectively avoid DNS pollution. But some native DoH providers may themselves be contaminated sometimes. In addition, some DoH services may be blocked by local network providers. Please choose the fastest and stablest DoH server.",
      "Awesome public DoH servers in Mainland China include alidns, geekdns, rubyfish, etc",
      "In taiwan area include quad101, etc",
      "USA: cloudflare, dns.google, etc",
      'Checklist：<a href="https://dnscrypt.info/public-servers" target="_blank">public-servers</a>',
      'Besides, setting up DoH service at your own native server is suggested and well-behaved in most cases <a href="https://github.com/facebookexperimental/doh-proxy" target="_blank">doh-proxy</a>. In this case, it is recommended to run the server(doh-proxy/doh-httpproxy) providing service and client(doh-stub) connecting to doh.opendns.com at the same time and connect them in series, because you can hardly find a server that is not polluted in a generally contaminated region.',
      "Optimally, place one or two lines above. The list will restore to default after saving with empty content."
    ]
  },
  dns: {
    title: "Configure DNS Server",
    dnsPriorityList: "Priority list of DNS Servers",
    messages: [
      "In the 'Prevent DNS Hijack Only' mode, the list is just the DNS configuration. In other modes, the first item in the list is the DNS server for querying Chinese mainland addresses.",
      "Please fill in the IP or domain of the DNS server (without port) in the list. Ports other than 53 are not supported.",
      "Optimally, place exact two lines above. The list will restore to default after saving with empty content."
    ]
  },
  egressPortWhitelist: {
    title: "Egress Port Whitelist",
    tcpPortWhitelist: "TCP Port Whitelist",
    udpPortWhitelist: "UDP Port Whitelist",
    messages: [
      "If v2rayA is setup on a server A which connected with a proxy server B, pay attention:",
      "Transparent proxy will force all TCP and UDP traffic to pass through proxy server B, where source IP address will be replaced with proxy B's. Moreover, if some clients send requests to server A that provides service, they will received responses from your proxy B's IP address weirdly, which is illegal.",
      "To resolve it, we need to add those service ports to whitelist so that not pass through proxy.For examples, ssh(22)、v2raya({v2rayaPort}).",
      "Obviously, if the server does not provide any service, you can skip configuring.",
      "Formatting：22 means port 22，20170:20172 means three ports 20170 to 20172."
    ]
  },
  configureServer: {
    title: "Configure Server | Server",
    servername: "Servername",
    port: "Port",
    forceTLS: "forcibly TLS on",
    noObfuscation: "No obfuscation",
    httpObfuscation: "Obfuscated as HTTP",
    srtpObfuscation: "Obfuscated as Video Calls (SRTP)",
    utpObfuscation: "Obfuscated as Bittorrent (uTP)",
    wechatVideoObfuscation: "Obfuscated as Wechat Video Calls",
    dtlsObfuscation: "Obfuscated as DTLS1.2 Packets",
    wireguardObfuscation: "Obfuscated as WireGuard Packets",
    hostObfuscation: "Host",
    pathObfuscation: "Path",
    password: "Password"
  },
  configureSubscription: {
    title: "Configure Subscription"
  },
  import: {
    message: "Input a vmess/ss/ssr/subscription address:"
  },
  delete: {
    title: "Confirm to DELETE",
    message:
      "Be sure to <b>DELETE</b> those servers/subscriptions? It is not reversible."
  },
  latency: {
    message:
      "Latency test used to cost one or several minutes. Wait patiently please."
  },
  version: {
    higherVersionNeeded:
      "This operation need higher version of v2rayA than {version}",
    v2rayInvalid: "v2ray-core may not be installed correctly"
  },
  about: `<p>v2rayA is a web GUI client of V2Ray. Frontend is built with Vue.js and backend is built with golang.</p>
          <p class="about-small">Default ports:</p>
          <p class="about-small">2017: v2rayA service port</p>
          <p class="about-small">20170: SOCKS protocol</p>
          <p class="about-small">20171: HTTP protocol</p>
          <p class="about-small">20172: HTTP protocol with "Rule of Splitting Traffic"</p>
          <p class="about-small">Other ports：</p>
          <p class="about-small">32345: tproxy, needed by transparent proxy </p>
          <p class="about-small">32346: port of plugins such as trojan, ssr and pingtunnel</p>
          <p>All data is stored in local instead of cloud. </p>
          <p>Problems found during use can be reported at <a href="https://github.com/v2rayA/v2rayA/issues">issues</a>.</p>`,
  axios: {
    messages: {
      optimizeBackend: "Adjust v2rayA service address？",
      noBackendFound:
        "Cannot find v2rayA at {url}. Make sure v2rayA is running at this address.",
      cannotCommunicate: [
        "Cannot communicate. If your service is running and ports open correctly, the reason may be that current browser does not allow https sites to access http resources, you can try using Chrome or switching to alternate http site.",
        "Cannot communicate. Firefox does not allow https sites to access http resources, you can try switching to alternate http site."
      ]
    },
    urls: {
      usage: "https://github.com/v2rayA/v2rayA/wiki/Usage"
    }
  },
  routingA: {
    messages: ["click the button 'Help&Manual' for help"]
  }
};
