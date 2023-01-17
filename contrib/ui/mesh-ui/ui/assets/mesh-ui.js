"use strict";
console.log("IE load fix");

var $ = id => document.getElementById(id)
var $$ = clazz => document.getElementsByClassName(clazz)
var ui = ui || {};
var riv = 'riv' in window ? window.riv : {};

ui.country_name={"ascension-island":"ac","andorra":"ad","united-arab-emirates":"ae","afghanistan":"af","antigua-and-barbuda":"ag","anguilla":"ai","albania":"al","armenia":"am","angola":"ao","antarctica":"aq","argentina":"ar","american-samoa":"as","austria":"at","australia":"au","aruba":"aw","aland-islands":"ax","azerbaijan":"az","bosnia-and-herzegovina":"ba","barbados":"bb","bangladesh":"bd","belgium":"be","burkina-faso":"bf","bulgaria":"bg","bahrain":"bh","burundi":"bi","benin":"bj","saint-barthélemy":"bl","bermuda":"bm","brunei-darussalam":"bn","bolivia":"bo","bonaire,-sint-eustatius-and-saba":"bq","brazil":"br","bahamas":"bs","bhutan":"bt","bouvet-island":"bv","botswana":"bw","belarus":"by","belize":"bz","canada":"ca","cocos-(keeling)-islands":"cc","democratic-republic-of-the-congo":"cd","central-european-free-trade-agreement":"cefta","central-african-republic":"cf","republic-of-the-congo":"cg","switzerland":"ch","côte-d'ivoire":"ci","cook-islands":"ck","chile":"cl","cameroon":"cm","china":"cn","colombia":"co","clipperton-island":"cp","costa-rica":"cr","cuba":"cu","cabo-verde":"cv","curaçao":"cw","christmas-island":"cx","cyprus":"cy","czech-republic":"cz","germany":"de","diego-garcia":"dg","djibouti":"dj","denmark":"dk","dominica":"dm","dominican-republic":"do","algeria":"dz","ceuta-&-melilla":"ea","ecuador":"ec","estonia":"ee","egypt":"eg","western-sahara":"eh","eritrea":"er","spain":"es","catalonia":"es-ct","galicia":"es-ga","ethiopia":"et","europe":"eu","finland":"fi","fiji":"fj","falkland-islands":"fk","federated-states-of-micronesia":"fm","faroe-islands":"fo","france":"fr","gabon":"ga","united-kingdom":"gb","england":"gb-eng","northern-ireland":"gb-nir","scotland":"gb-sct","wales":"gb-wls","grenada":"gd","georgia":"ge","french-guiana":"gf","guernsey":"gg","ghana":"gh","gibraltar":"gi","greenland":"gl","gambia":"gm","guinea":"gn","guadeloupe":"gp","equatorial-guinea":"gq","greece":"gr","south-georgia-and-the-south-sandwich-islands":"gs","guatemala":"gt","guam":"gu","guinea-bissau":"gw","guyana":"gy","hong-kong":"hk","heard-island-and-mcdonald-islands":"hm","honduras":"hn","croatia":"hr","haiti":"ht","hungary":"hu","canary-islands":"ic","indonesia":"id","ireland":"ie","israel":"il","isle-of-man":"im","india":"in","british-indian-ocean-territory":"io","iraq":"iq","iran":"ir","iceland":"is","italy":"it","jersey":"je","jamaica":"jm","jordan":"jo","japan":"jp","kenya":"ke","kyrgyzstan":"kg","cambodia":"kh","kiribati":"ki","comoros":"km","saint-kitts-and-nevis":"kn","north-korea":"kp","south-korea":"kr","kuwait":"kw","cayman-islands":"ky","kazakhstan":"kz","laos":"la","lebanon":"lb","saint-lucia":"lc","liechtenstein":"li","sri-lanka":"lk","liberia":"lr","lesotho":"ls","lithuania":"lt","luxembourg":"lu","latvia":"lv","libya":"ly","morocco":"ma","monaco":"mc","moldova":"md","montenegro":"me","saint-martin":"mf","madagascar":"mg","marshall-islands":"mh","north-macedonia":"mk","mali":"ml","myanmar":"mm","mongolia":"mn","macau":"mo","northern-mariana-islands":"mp","martinique":"mq","mauritania":"mr","montserrat":"ms","malta":"mt","mauritius":"mu","maldives":"mv","malawi":"mw","mexico":"mx","malaysia":"my","mozambique":"mz","namibia":"na","new-caledonia":"nc","niger":"ne","norfolk-island":"nf","nigeria":"ng","nicaragua":"ni","netherlands":"nl","norway":"no","nepal":"np","nauru":"nr","niue":"nu","new-zealand":"nz","oman":"om","panama":"pa","peru":"pe","french-polynesia":"pf","papua-new-guinea":"pg","philippines":"ph","pakistan":"pk","poland":"pl","saint-pierre-and-miquelon":"pm","pitcairn":"pn","puerto-rico":"pr","state-of-palestine":"ps","portugal":"pt","palau":"pw","paraguay":"py","qatar":"qa","réunion":"re","romania":"ro","serbia":"rs","russia":"ru","rwanda":"rw","saudi-arabia":"sa","solomon-islands":"sb","seychelles":"sc","sudan":"sd","sweden":"se","singapore":"sg","saint-helena,-ascension-and-tristan-da-cunha":"sh","slovenia":"si","svalbard-and-jan-mayen":"sj","slovakia":"sk","sierra-leone":"sl","san-marino":"sm","senegal":"sn","somalia":"so","suriname":"sr","south-sudan":"ss","sao-tome-and-principe":"st","el-salvador":"sv","sint-maarten":"sx","syria":"sy","eswatini":"sz","tristan-da-cunha":"ta","turks-and-caicos-islands":"tc","chad":"td","french-southern-territories":"tf","togo":"tg","thailand":"th","tajikistan":"tj","tokelau":"tk","timor-leste":"tl","turkmenistan":"tm","tunisia":"tn","tonga":"to","turkey":"tr","trinidad-and-tobago":"tt","tuvalu":"tv","taiwan":"tw","tanzania":"tz","ukraine":"ua","uganda":"ug","united-states-minor-outlying-islands":"um","united-nations":"un","united-states-of-america":"us","uruguay":"uy","uzbekistan":"uz","holy-see":"va","saint-vincent-and-the-grenadines":"vc","venezuela":"ve","virgin-islands-(british)":"vg","virgin-islands-(u.s.)":"vi","vietnam":"vn","vanuatu":"vu","wallis-and-futuna":"wf","samoa":"ws","kosovo":"xk","unknown":"xx","yemen":"ye","mayotte":"yt","south-africa":"za","zambia":"zm","zimbabwe":"zw"};
ui.country_name["united-states"] = "us";
ui.country_name["czechia"] = "cz";

function setHealth(d) {
  var style = window.getComputedStyle($("peer_list"));
  if(style.display === 'none') {
    return;
  }

  let trs = $("peer_list").rows;
  //remove duplications
  for (let i=0; i < trs.length; ++i) {
    if(trs[i].cells[1].textContent == d.peer) {
      tb.deleteRow(1);
      break;
    }
  }
  // creates a table row
  var row = document.createElement("tr");
  var imgElement = document.createElement("td");
  var peerAddress = document.createElement("td");
  peerAddress.innerText = d.peer;
  peerAddress.className = "all_peers_url";
  var peerPing = document.createElement("td");
  var peerPingTime = document.createElement("td");
  var peerSelect = document.createElement("td");
  var chk = document.createElement('input');
  chk.setAttribute('type', 'checkbox');
  chk.checked = ui.connectedPeersAddress.indexOf(d.peer) >= 0;
  chk.setAttribute('id', "peer-" + d.peer);  
  peerSelect.appendChild(chk);

  row.appendChild(imgElement);
  row.appendChild(peerAddress);
  row.appendChild(peerPing);
  row.appendChild(peerPingTime);
  row.appendChild(peerSelect);

  if ((d.peer in ui.peers_country) && (ui.peers_country[d.peer] in ui.country_name))
    imgElement.className = "big-flag fi fi-" + ui.country_name[ui.peers_country[d.peer]];
  else if(d.country_short)
    imgElement.className = "big-flag fi fi-" + d.country_short.toLowerCase();
  else
    imgElement.className = "fas fa-thin fa-share-nodes";

  if (!("ping" in d)) {
    peerAddress.style.color = "rgba(250,250,250,.5)";
  } else {
    peerPing.innerText = d.ping.toFixed(0);
    peerPingTime.appendChild(document.createTextNode("ms"));
  }
  
  //sort table
  insertRowToOrderPos($("peer_list"), 2, row)
}

function cmpTime(a, b) {
  a = a.textContent.trim() || "999999";
  b = b.textContent.trim() || "999999";
  return a.localeCompare(b, 'en', { numeric: true })
}

function insertRowToOrderPos(tb, col, row) {
  let tr = tb.rows;

  var i = 0;
  for (; i < tr.length && cmpTime(row.cells[col], tr[i].cells[col]) >= 0; ++i);
  if (i < tr.length) {
    tb.insertBefore(row, tr[i]);
  } else {
    tb.appendChild(row);
  }
}

function openTab(element, tabName) {
  // Declare all variables
  var i, tabContent, tabLinks;

  // Get all elements with class="content" and hide them
  tabContent = $$("tab here");
  for (i = 0; i < tabContent.length; i++) {
    tabContent[i].className = "tab here is-hidden";
  }

  // Get all elements with class="tab" and remove the class "is-active"
  tabLinks = $$("tab is-active");
  for (i = 0; i < tabLinks.length; i++) {
    tabLinks[i].className = "tab";
  }

  // Show the current tab, and add an "is-active" class to the button that opened the tab
  $(tabName).className = "tab here";
  element.parentElement.className = "tab is-active";
  //refreshRecordsList();
}

function copy2clipboard(text) {
  var textArea = document.createElement("textarea");
  textArea.style.position = 'fixed';
  textArea.style.top = 0;
  textArea.style.left = 0;

  // Ensure it has a small width and height. Setting to 1px / 1em
  // doesn't work as this gives a negative w/h on some browsers.
  textArea.style.width = '2em';
  textArea.style.height = '2em';

  // We don't need padding, reducing the size if it does flash render.
  textArea.style.padding = 0;

  // Clean up any borders.
  textArea.style.border = 'none';
  textArea.style.outline = 'none';
  textArea.style.boxShadow = 'none';

  // Avoid flash of the white box if rendered for any reason.
  textArea.style.background = 'transparent';
  textArea.value = text;

  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();
  try {
    var successful = document.execCommand('copy');
    var msg = successful ? 'successful' : 'unsuccessful';
    console.log('Copying text command was ' + msg);
  } catch (err) {
    console.log('Oops, unable to copy');
  }
  document.body.removeChild(textArea);
  showInfo('value copied successfully!');
}

function showInfo(text) {
  var info = $("notification_info");
  var message = $("info_text");
  message.innerHTML = text;

  info.className = "notification is-primary";
  var button = $("info_close");
  button.onclick = function () {
    message.value = "";
    info.className = "notification is-primary is-hidden";
  };
  setTimeout(button.onclick, 2000);
}

function showError(text) {
  var info = $("notification_error");
  var message = $("error_text");
  message.innerHTML = text;

  info.className = "notification is-danger";
  var button = $("info_close");
  button.onclick = function () {
    message.value = "";
    info.className = "notification is-danger is-hidden";
  };
}

function showWindow() {
  var info = $("notification_window");
  var message = $("info_window");

  info.classList.remove("is-hidden");
  var button_info_close = $("info_win_close");
  button_info_close.onclick = function () {
    info.classList.add("is-hidden");
    $("peer_list").innerHTML = "";
  };
  var button_window_close = $("window_close");
  button_window_close.onclick = function () {
    info.classList.add("is-hidden");
    $("peer_list").innerHTML = "";
  };
  var button_window_save = $("window_save");
  button_window_save.onclick = function () {
    info.classList.add("is-hidden");
    //todo save peers
    var peers = document.querySelectorAll('*[id^="peer-"]');
    var peer_list = [];
    for (var i = 0; i < peers.length; ++i) {
      var p = peers[i];
      if (p.checked) {
        var peerURL = p.parentElement.parentElement.children[1].innerText;
        peer_list.push(peerURL);
      }
    }
    fetch('api/peers', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Riv-Save-Config': 'true',
        },
        body: JSON.stringify(peer_list.map(x => {return {"url": x}})),
      })
      .catch((error) => {
        console.error('Error:', error);
      });    
    $("peer_list").innerHTML = "";
  };
}

function togglePrivKeyVisibility() {
  if (this.classList.contains("fa-eye")) {
    this.classList.remove("fa-eye");
    this.classList.add("fa-eye-slash");
    $("priv_key_visible").innerHTML = $("priv_key").innerHTML;
  } else {
    this.classList.remove("fa-eye-slash");
    this.classList.add("fa-eye");
    $("priv_key_visible").innerHTML = "••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••";
  }
}

function humanReadableSpeed(speed) {
  if (speed < 0) return "? B/s";
  var i = speed < 1 ? 0 : Math.floor(Math.log(speed) / Math.log(1024));
  var val = speed / Math.pow(1024, i);
  var fixed = 2;
  if((val.toFixed() * 1) > 99) {
    i+=1;
    val /= 1024
  } else if((val.toFixed() * 1) > 9) {
    fixed = 1;
  }
  return val.toFixed(fixed) + ' ' + ['B/s', 'kB/s', 'MB/s', 'GB/s', 'TB/s'][i];
}

ui.showAllPeers = async () => {
  try {
    let response = await fetch('api/publicpeers')
    let peerList = await response.json();
    showWindow();
    ui.peers_country = Object.keys(peerList).flatMap(country => Object.keys(peerList[country]).map(peer => {let r={}; r[peer] = country.replace(".md", ""); return r}));
    ui.peers_country = ui.peers_country.reduce(((r, c) => Object.assign(r, c)), {})
    const peers = Object.values(peerList).flatMap(x => Object.keys(x));
        //start peers test
    await fetch('api/health', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(peers)
        });
  } catch(e) {
    console.error('Error:', e);
  }
}

ui.getConnectedPeers = () =>
  fetch('api/peers')
    .then((response) => response.json())

ui.updateConnectedPeersHandler = (peers) => {
  ui.updateStatus(peers);
  $("peers").innerText = "";
  ui.connectedPeersAddress = peers.map(peer => peer.remote);
  if(peers) {
    const regexStrip = /%[^\]]*/gm;
    peers.forEach(peer => {
      let row = $("peers").appendChild(document.createElement('div'));
      row.className = "overflow-ellipsis"
      let flag =  row.appendChild(document.createElement("span"));
      if(peer.multicast || !peer.country_short)
        flag.className = "fas fa-thin fa-share-nodes peer-connected-fl";
      else
        flag.className = "fi fi-" + peer.country_short.toLowerCase() + " peer-connected-fl";
      row.append(peer.remote.replace(regexStrip, ""));
    });
  }
}

ui.updateStatus = peers => {
  let status = "st-error";
  if(peers) {
    if(peers.length) {
      const isNonMulticastExists = peers.filter(peer => !peer.multicast).length;
      status = isNonMulticastExists ? "st-multicast" : "st-connected";
    } else {
      status = "st-connecting"
    }
  }
  Array.from($$("status")).forEach(node => node.classList.add("is-hidden"));
  $(status).classList.remove("is-hidden");
}

ui.updateSpeed = peers => {
  if(peers) {
    let rsbytes = {"bytes_recvd": peers.reduce((acc, peer) => acc + peer.bytes_recvd, 0),
                   "bytes_sent":  peers.reduce((acc, peer) => acc + peer.bytes_sent, 0),
                   "timestamp": Date.now()};
    if("_rsbytes" in ui) {
      $("dn_speed").innerText = humanReadableSpeed((rsbytes.bytes_recvd - ui._rsbytes.bytes_recvd) * 1000 / (rsbytes.timestamp - ui._rsbytes.timestamp));
      $("up_speed").innerText = humanReadableSpeed((rsbytes.bytes_sent - ui._rsbytes.bytes_sent) * 1000 / (rsbytes.timestamp - ui._rsbytes.timestamp));
    }
    ui._rsbytes = rsbytes;
  } else {
    delete ui._rsbytes;
    $("dn_speed").innerText = humanReadableSpeed(-1);
    $("up_speed").innerText = humanReadableSpeed(-1);
  }
}

ui.updateConnectedPeers = () =>
  ui.getConnectedPeers()
    .then(peers => ui.updateConnectedPeersHandler(peers))
    .catch((error) => {
      ui.updateConnectedPeersHandler();
      $("peers").innerText = error.message;
    });

ui.getSelfInfo = () =>
  fetch('api/self')
    .then((response) => {
      if (response.status === 401) {
        let status="st-error"
        Array.from($$("status")).forEach(node => node.classList.add("is-hidden"));
        $(status).classList.remove("is-hidden");
        $("logoutButton").classList.add('is-active');
        $("logoutButton").classList.remove('is-hidden');
        response.text().then(text => {
          if (riv.useAuthNASRichScreen) {
            $("login").classList.remove('is-hidden');
            $("login").classList.add('is-active');
            $("username").text(riv.getNasUser());
          } else {
            showError(text);
            $("progress").classList.remove('is-active');
            $("progress").classList.add('is-hidden');
          }
        });
      } else {
        return response.json()
      }
      if (riv.useAuthNASRichScreen){
        $("logoutButton").classList.remove('is-hidden');
        $("logoutButton").classList.add('is-active');
      }
    });

ui.updateSelfInfo = () =>
  ui.getSelfInfo()
    .then((info) => {
      $("ipv6").innerText = info.address;
      $("subnet").innerText = info.subnet;
      $("coordinates").innerText = ''.concat('[',info.coords.join(' '),']');
      $("pub_key").innerText = info.key;
      $("priv_key").innerText = info.private_key;
      $("ipv6").innerText = info.address;
      $("version").innerText = info.build_version;
    }).catch((error) => {
      $("ipv6").innerText = error.message;
    });

var nasLoginSuccess = function () {
  $("login").classList.remove('is-active');
  $("login").classList.add('is-hidden');
  $("progress").classList.remove('is-active');
  $("progress").classList.add('is-hidden');
};

var nasLoginFailure = function (message) {
  //Show notification: Username/password is wrong
  if (typeof message !== 'undefined') {
    showError(message);
  } else {
    showError("Incorrect username or password");
  }
  $('password').val('');
  $("login").classList.add('is-active');
  $("login").classList.remove('is-hidden');
  $("progress").classList.remove('is-active');
  $("progress").classList.add('is-hidden');
};

ui.handleLogin = () =>
$("loginButton").addEventListener("click", function (e) {
  e.preventDefault();
  $('username').value = $('username').value.trim();
  if ($('username').value.length === 0 || $('password').value.trim().length === 0) {
    return;
  }
  $("login").classList.add('is-hidden');
  $("login").classList.remove('is-active');
  $("progress").classList.remove('is-hidden');
  $("progress").classList.add('is-active');
  riv.nasLoginCall(nasLoginSuccess, nasLoginFailure);
});

function main() {

  window.addEventListener("load", () => {
    $("showAllPeersBtn").addEventListener("click", ui.showAllPeers);
    Array.from($$("target_new_window")).forEach(a => {
      a.addEventListener("click", (event)=>{
        event.preventDefault();
        window.open(new URL(new URL(a.href).hash.substring(1), location.origin).href);
      });
    });

    $("logout").addEventListener("click", (event)=> {
      event.preventDefault();
      riv.nasLogoutCall();
      window.location.reload();
    });

    ui.handleLogin();

    ui.updateSelfInfo();

    ui.updateConnectedPeers();

    ui.sse = new EventSource('api/sse');

    ui.sse.addEventListener("health", (e) => {
      setHealth(JSON.parse(e.data));
    })
    
    ui.sse.addEventListener("peers", (e) => {
      ui.updateConnectedPeersHandler(JSON.parse(e.data));
    })
    
    ui.sse.addEventListener("rxtx", (e) => {
      ui.updateSpeed(JSON.parse(e.data));
    })
    
    ui.sse.addEventListener("coord", (e) => {
      let coords = JSON.parse(e.data);
      $("coordinates").innerText = ''.concat('[',coords.join(' '),']');
    })
    
  });
}

main();
