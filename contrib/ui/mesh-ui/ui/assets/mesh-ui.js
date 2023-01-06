var $ = id => document.getElementById(id)
var $$ = clazz => document.getElementsByClassName(clazz)
var ui = ui || {};

ui.country_name=[{"Ascension Island":"ac"},{"Andorra":"ad"},{"United Arab Emirates":"ae"},{"Afghanistan":"af"},{"Antigua and Barbuda":"ag"},{"Anguilla":"ai"},{"Albania":"al"},{"Armenia":"am"},{"Angola":"ao"},{"Antarctica":"aq"},{"Argentina":"ar"},{"American Samoa":"as"},{"Austria":"at"},{"Australia":"au"},{"Aruba":"aw"},{"Aland Islands":"ax"},{"Azerbaijan":"az"},{"Bosnia and Herzegovina":"ba"},{"Barbados":"bb"},{"Bangladesh":"bd"},{"Belgium":"be"},{"Burkina Faso":"bf"},{"Bulgaria":"bg"},{"Bahrain":"bh"},{"Burundi":"bi"},{"Benin":"bj"},{"Saint Barthélemy":"bl"},{"Bermuda":"bm"},{"Brunei Darussalam":"bn"},{"Bolivia":"bo"},{"Bonaire, Sint Eustatius and Saba":"bq"},{"Brazil":"br"},{"Bahamas":"bs"},{"Bhutan":"bt"},{"Bouvet Island":"bv"},{"Botswana":"bw"},{"Belarus":"by"},{"Belize":"bz"},{"Canada":"ca"},{"Cocos (Keeling) Islands":"cc"},{"Democratic Republic of the Congo":"cd"},{"Central European Free Trade Agreement":"cefta"},{"Central African Republic":"cf"},{"Republic of the Congo":"cg"},{"Switzerland":"ch"},{"Côte d'Ivoire":"ci"},{"Cook Islands":"ck"},{"Chile":"cl"},{"Cameroon":"cm"},{"China":"cn"},{"Colombia":"co"},{"Clipperton Island":"cp"},{"Costa Rica":"cr"},{"Cuba":"cu"},{"Cabo Verde":"cv"},{"Curaçao":"cw"},{"Christmas Island":"cx"},{"Cyprus":"cy"},{"Czech Republic":"cz"},{"Germany":"de"},{"Diego Garcia":"dg"},{"Djibouti":"dj"},{"Denmark":"dk"},{"Dominica":"dm"},{"Dominican Republic":"do"},{"Algeria":"dz"},{"Ceuta & Melilla":"ea"},{"Ecuador":"ec"},{"Estonia":"ee"},{"Egypt":"eg"},{"Western Sahara":"eh"},{"Eritrea":"er"},{"Spain":"es"},{"Catalonia":"es-ct"},{"Galicia":"es-ga"},{"Ethiopia":"et"},{"Europe":"eu"},{"Finland":"fi"},{"Fiji":"fj"},{"Falkland Islands":"fk"},{"Federated States of Micronesia":"fm"},{"Faroe Islands":"fo"},{"France":"fr"},{"Gabon":"ga"},{"United Kingdom":"gb"},{"England":"gb-eng"},{"Northern Ireland":"gb-nir"},{"Scotland":"gb-sct"},{"Wales":"gb-wls"},{"Grenada":"gd"},{"Georgia":"ge"},{"French Guiana":"gf"},{"Guernsey":"gg"},{"Ghana":"gh"},{"Gibraltar":"gi"},{"Greenland":"gl"},{"Gambia":"gm"},{"Guinea":"gn"},{"Guadeloupe":"gp"},{"Equatorial Guinea":"gq"},{"Greece":"gr"},{"South Georgia and the South Sandwich Islands":"gs"},{"Guatemala":"gt"},{"Guam":"gu"},{"Guinea-Bissau":"gw"},{"Guyana":"gy"},{"Hong Kong":"hk"},{"Heard Island and McDonald Islands":"hm"},{"Honduras":"hn"},{"Croatia":"hr"},{"Haiti":"ht"},{"Hungary":"hu"},{"Canary Islands":"ic"},{"Indonesia":"id"},{"Ireland":"ie"},{"Israel":"il"},{"Isle of Man":"im"},{"India":"in"},{"British Indian Ocean Territory":"io"},{"Iraq":"iq"},{"Iran":"ir"},{"Iceland":"is"},{"Italy":"it"},{"Jersey":"je"},{"Jamaica":"jm"},{"Jordan":"jo"},{"Japan":"jp"},{"Kenya":"ke"},{"Kyrgyzstan":"kg"},{"Cambodia":"kh"},{"Kiribati":"ki"},{"Comoros":"km"},{"Saint Kitts and Nevis":"kn"},{"North Korea":"kp"},{"South Korea":"kr"},{"Kuwait":"kw"},{"Cayman Islands":"ky"},{"Kazakhstan":"kz"},{"Laos":"la"},{"Lebanon":"lb"},{"Saint Lucia":"lc"},{"Liechtenstein":"li"},{"Sri Lanka":"lk"},{"Liberia":"lr"},{"Lesotho":"ls"},{"Lithuania":"lt"},{"Luxembourg":"lu"},{"Latvia":"lv"},{"Libya":"ly"},{"Morocco":"ma"},{"Monaco":"mc"},{"Moldova":"md"},{"Montenegro":"me"},{"Saint Martin":"mf"},{"Madagascar":"mg"},{"Marshall Islands":"mh"},{"North Macedonia":"mk"},{"Mali":"ml"},{"Myanmar":"mm"},{"Mongolia":"mn"},{"Macau":"mo"},{"Northern Mariana Islands":"mp"},{"Martinique":"mq"},{"Mauritania":"mr"},{"Montserrat":"ms"},{"Malta":"mt"},{"Mauritius":"mu"},{"Maldives":"mv"},{"Malawi":"mw"},{"Mexico":"mx"},{"Malaysia":"my"},{"Mozambique":"mz"},{"Namibia":"na"},{"New Caledonia":"nc"},{"Niger":"ne"},{"Norfolk Island":"nf"},{"Nigeria":"ng"},{"Nicaragua":"ni"},{"Netherlands":"nl"},{"Norway":"no"},{"Nepal":"np"},{"Nauru":"nr"},{"Niue":"nu"},{"New Zealand":"nz"},{"Oman":"om"},{"Panama":"pa"},{"Peru":"pe"},{"French Polynesia":"pf"},{"Papua New Guinea":"pg"},{"Philippines":"ph"},{"Pakistan":"pk"},{"Poland":"pl"},{"Saint Pierre and Miquelon":"pm"},{"Pitcairn":"pn"},{"Puerto Rico":"pr"},{"State of Palestine":"ps"},{"Portugal":"pt"},{"Palau":"pw"},{"Paraguay":"py"},{"Qatar":"qa"},{"Réunion":"re"},{"Romania":"ro"},{"Serbia":"rs"},{"Russia":"ru"},{"Rwanda":"rw"},{"Saudi Arabia":"sa"},{"Solomon Islands":"sb"},{"Seychelles":"sc"},{"Sudan":"sd"},{"Sweden":"se"},{"Singapore":"sg"},{"Saint Helena, Ascension and Tristan da Cunha":"sh"},{"Slovenia":"si"},{"Svalbard and Jan Mayen":"sj"},{"Slovakia":"sk"},{"Sierra Leone":"sl"},{"San Marino":"sm"},{"Senegal":"sn"},{"Somalia":"so"},{"Suriname":"sr"},{"South Sudan":"ss"},{"Sao Tome and Principe":"st"},{"El Salvador":"sv"},{"Sint Maarten":"sx"},{"Syria":"sy"},{"Eswatini":"sz"},{"Tristan da Cunha":"ta"},{"Turks and Caicos Islands":"tc"},{"Chad":"td"},{"French Southern Territories":"tf"},{"Togo":"tg"},{"Thailand":"th"},{"Tajikistan":"tj"},{"Tokelau":"tk"},{"Timor-Leste":"tl"},{"Turkmenistan":"tm"},{"Tunisia":"tn"},{"Tonga":"to"},{"Turkey":"tr"},{"Trinidad and Tobago":"tt"},{"Tuvalu":"tv"},{"Taiwan":"tw"},{"Tanzania":"tz"},{"Ukraine":"ua"},{"Uganda":"ug"},{"United States Minor Outlying Islands":"um"},{"United Nations":"un"},{"United States of America":"us"},{"Uruguay":"uy"},{"Uzbekistan":"uz"},{"Holy See":"va"},{"Saint Vincent and the Grenadines":"vc"},{"Venezuela":"ve"},{"Virgin Islands (British)":"vg"},{"Virgin Islands (U.S.)":"vi"},{"Vietnam":"vn"},{"Vanuatu":"vu"},{"Wallis and Futuna":"wf"},{"Samoa":"ws"},{"Kosovo":"xk"},{"Unknown":"xx"},{"Yemen":"ye"},{"Mayotte":"yt"},{"South Africa":"za"},{"Zambia":"zm"},{"Zimbabwe":"zw"}];

function setHealth(d) {
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
  peerSelect.appendChild(chk);

  row.appendChild(imgElement);
  row.appendChild(peerAddress);
  row.appendChild(peerPing);
  row.appendChild(peerPingTime);
  row.appendChild(peerSelect);

  if(d.country_short)
    imgElement.className = "big-flag fi fi-" + d.country_short.toLowerCase();
  else if (d.peer in ui.peers_country)
    imgElement.className = "big-flag fi fi-" + ui.peers_country[ui.country_name[d.peer]].toLowerCase();
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
    .then((response) => response.json())

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

function main() {

  window.addEventListener("load", () => {
    $("showAllPeersBtn").addEventListener("click", ui.showAllPeers);

    ui.updateConnectedPeers();

    ui.updateSelfInfo();

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
