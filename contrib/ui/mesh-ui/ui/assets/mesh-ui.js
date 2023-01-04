var $ = id => document.getElementById(id)
var $$ = clazz => document.getElementsByClassName(clazz)

function setHealth(d) {
  // creates a table row
  var row = document.createElement("tr");
  var imgElement = document.createElement("td");
  var peerAddress = document.createElement("td");
  peerAddress.innerText = d.peer;
  var peerPing = document.createElement("td");
  peerPing.setAttribute('id', d.peer);
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
  return a.textContent.trim() === "" ? 1 : (a.textContent.trim() // using `.textContent.trim()` for test
    .localeCompare(b.textContent.trim(), 'en', { numeric: true }))
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

var ui = ui || {};


ui.showAllPeers = async () => {
  try {
    let response = await fetch('https://map.rivchain.org/rest/peers.json')
    let peerList = await response.json();
    showWindow();
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

    ui.sse = new EventSource('/api/sse');

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
