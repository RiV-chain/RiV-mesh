"use strict";
console.log("IE load fix");

function _typeof(obj) { "@babel/helpers - typeof"; return _typeof = "function" == typeof Symbol && "symbol" == typeof Symbol.iterator ? function (obj) { return typeof obj; } : function (obj) { return obj && "function" == typeof Symbol && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }, _typeof(obj); }
function _regeneratorRuntime() { "use strict"; /*! regenerator-runtime -- Copyright (c) 2014-present, Facebook, Inc. -- license (MIT): https://github.com/facebook/regenerator/blob/main/LICENSE */ _regeneratorRuntime = function _regeneratorRuntime() { return exports; }; var exports = {}, Op = Object.prototype, hasOwn = Op.hasOwnProperty, defineProperty = Object.defineProperty || function (obj, key, desc) { obj[key] = desc.value; }, $Symbol = "function" == typeof Symbol ? Symbol : {}, iteratorSymbol = $Symbol.iterator || "@@iterator", asyncIteratorSymbol = $Symbol.asyncIterator || "@@asyncIterator", toStringTagSymbol = $Symbol.toStringTag || "@@toStringTag"; function define(obj, key, value) { return Object.defineProperty(obj, key, { value: value, enumerable: !0, configurable: !0, writable: !0 }), obj[key]; } try { define({}, ""); } catch (err) { define = function define(obj, key, value) { return obj[key] = value; }; } function wrap(innerFn, outerFn, self, tryLocsList) { var protoGenerator = outerFn && outerFn.prototype instanceof Generator ? outerFn : Generator, generator = Object.create(protoGenerator.prototype), context = new Context(tryLocsList || []); return defineProperty(generator, "_invoke", { value: makeInvokeMethod(innerFn, self, context) }), generator; } function tryCatch(fn, obj, arg) { try { return { type: "normal", arg: fn.call(obj, arg) }; } catch (err) { return { type: "throw", arg: err }; } } exports.wrap = wrap; var ContinueSentinel = {}; function Generator() {} function GeneratorFunction() {} function GeneratorFunctionPrototype() {} var IteratorPrototype = {}; define(IteratorPrototype, iteratorSymbol, function () { return this; }); var getProto = Object.getPrototypeOf, NativeIteratorPrototype = getProto && getProto(getProto(values([]))); NativeIteratorPrototype && NativeIteratorPrototype !== Op && hasOwn.call(NativeIteratorPrototype, iteratorSymbol) && (IteratorPrototype = NativeIteratorPrototype); var Gp = GeneratorFunctionPrototype.prototype = Generator.prototype = Object.create(IteratorPrototype); function defineIteratorMethods(prototype) { ["next", "throw", "return"].forEach(function (method) { define(prototype, method, function (arg) { return this._invoke(method, arg); }); }); } function AsyncIterator(generator, PromiseImpl) { function invoke(method, arg, resolve, reject) { var record = tryCatch(generator[method], generator, arg); if ("throw" !== record.type) { var result = record.arg, value = result.value; return value && "object" == _typeof(value) && hasOwn.call(value, "__await") ? PromiseImpl.resolve(value.__await).then(function (value) { invoke("next", value, resolve, reject); }, function (err) { invoke("throw", err, resolve, reject); }) : PromiseImpl.resolve(value).then(function (unwrapped) { result.value = unwrapped, resolve(result); }, function (error) { return invoke("throw", error, resolve, reject); }); } reject(record.arg); } var previousPromise; defineProperty(this, "_invoke", { value: function value(method, arg) { function callInvokeWithMethodAndArg() { return new PromiseImpl(function (resolve, reject) { invoke(method, arg, resolve, reject); }); } return previousPromise = previousPromise ? previousPromise.then(callInvokeWithMethodAndArg, callInvokeWithMethodAndArg) : callInvokeWithMethodAndArg(); } }); } function makeInvokeMethod(innerFn, self, context) { var state = "suspendedStart"; return function (method, arg) { if ("executing" === state) throw new Error("Generator is already running"); if ("completed" === state) { if ("throw" === method) throw arg; return doneResult(); } for (context.method = method, context.arg = arg;;) { var delegate = context.delegate; if (delegate) { var delegateResult = maybeInvokeDelegate(delegate, context); if (delegateResult) { if (delegateResult === ContinueSentinel) continue; return delegateResult; } } if ("next" === context.method) context.sent = context._sent = context.arg;else if ("throw" === context.method) { if ("suspendedStart" === state) throw state = "completed", context.arg; context.dispatchException(context.arg); } else "return" === context.method && context.abrupt("return", context.arg); state = "executing"; var record = tryCatch(innerFn, self, context); if ("normal" === record.type) { if (state = context.done ? "completed" : "suspendedYield", record.arg === ContinueSentinel) continue; return { value: record.arg, done: context.done }; } "throw" === record.type && (state = "completed", context.method = "throw", context.arg = record.arg); } }; } function maybeInvokeDelegate(delegate, context) { var methodName = context.method, method = delegate.iterator[methodName]; if (undefined === method) return context.delegate = null, "throw" === methodName && delegate.iterator.return && (context.method = "return", context.arg = undefined, maybeInvokeDelegate(delegate, context), "throw" === context.method) || "return" !== methodName && (context.method = "throw", context.arg = new TypeError("The iterator does not provide a '" + methodName + "' method")), ContinueSentinel; var record = tryCatch(method, delegate.iterator, context.arg); if ("throw" === record.type) return context.method = "throw", context.arg = record.arg, context.delegate = null, ContinueSentinel; var info = record.arg; return info ? info.done ? (context[delegate.resultName] = info.value, context.next = delegate.nextLoc, "return" !== context.method && (context.method = "next", context.arg = undefined), context.delegate = null, ContinueSentinel) : info : (context.method = "throw", context.arg = new TypeError("iterator result is not an object"), context.delegate = null, ContinueSentinel); } function pushTryEntry(locs) { var entry = { tryLoc: locs[0] }; 1 in locs && (entry.catchLoc = locs[1]), 2 in locs && (entry.finallyLoc = locs[2], entry.afterLoc = locs[3]), this.tryEntries.push(entry); } function resetTryEntry(entry) { var record = entry.completion || {}; record.type = "normal", delete record.arg, entry.completion = record; } function Context(tryLocsList) { this.tryEntries = [{ tryLoc: "root" }], tryLocsList.forEach(pushTryEntry, this), this.reset(!0); } function values(iterable) { if (iterable) { var iteratorMethod = iterable[iteratorSymbol]; if (iteratorMethod) return iteratorMethod.call(iterable); if ("function" == typeof iterable.next) return iterable; if (!isNaN(iterable.length)) { var i = -1, next = function next() { for (; ++i < iterable.length;) if (hasOwn.call(iterable, i)) return next.value = iterable[i], next.done = !1, next; return next.value = undefined, next.done = !0, next; }; return next.next = next; } } return { next: doneResult }; } function doneResult() { return { value: undefined, done: !0 }; } return GeneratorFunction.prototype = GeneratorFunctionPrototype, defineProperty(Gp, "constructor", { value: GeneratorFunctionPrototype, configurable: !0 }), defineProperty(GeneratorFunctionPrototype, "constructor", { value: GeneratorFunction, configurable: !0 }), GeneratorFunction.displayName = define(GeneratorFunctionPrototype, toStringTagSymbol, "GeneratorFunction"), exports.isGeneratorFunction = function (genFun) { var ctor = "function" == typeof genFun && genFun.constructor; return !!ctor && (ctor === GeneratorFunction || "GeneratorFunction" === (ctor.displayName || ctor.name)); }, exports.mark = function (genFun) { return Object.setPrototypeOf ? Object.setPrototypeOf(genFun, GeneratorFunctionPrototype) : (genFun.__proto__ = GeneratorFunctionPrototype, define(genFun, toStringTagSymbol, "GeneratorFunction")), genFun.prototype = Object.create(Gp), genFun; }, exports.awrap = function (arg) { return { __await: arg }; }, defineIteratorMethods(AsyncIterator.prototype), define(AsyncIterator.prototype, asyncIteratorSymbol, function () { return this; }), exports.AsyncIterator = AsyncIterator, exports.async = function (innerFn, outerFn, self, tryLocsList, PromiseImpl) { void 0 === PromiseImpl && (PromiseImpl = Promise); var iter = new AsyncIterator(wrap(innerFn, outerFn, self, tryLocsList), PromiseImpl); return exports.isGeneratorFunction(outerFn) ? iter : iter.next().then(function (result) { return result.done ? result.value : iter.next(); }); }, defineIteratorMethods(Gp), define(Gp, toStringTagSymbol, "Generator"), define(Gp, iteratorSymbol, function () { return this; }), define(Gp, "toString", function () { return "[object Generator]"; }), exports.keys = function (val) { var object = Object(val), keys = []; for (var key in object) keys.push(key); return keys.reverse(), function next() { for (; keys.length;) { var key = keys.pop(); if (key in object) return next.value = key, next.done = !1, next; } return next.done = !0, next; }; }, exports.values = values, Context.prototype = { constructor: Context, reset: function reset(skipTempReset) { if (this.prev = 0, this.next = 0, this.sent = this._sent = undefined, this.done = !1, this.delegate = null, this.method = "next", this.arg = undefined, this.tryEntries.forEach(resetTryEntry), !skipTempReset) for (var name in this) "t" === name.charAt(0) && hasOwn.call(this, name) && !isNaN(+name.slice(1)) && (this[name] = undefined); }, stop: function stop() { this.done = !0; var rootRecord = this.tryEntries[0].completion; if ("throw" === rootRecord.type) throw rootRecord.arg; return this.rval; }, dispatchException: function dispatchException(exception) { if (this.done) throw exception; var context = this; function handle(loc, caught) { return record.type = "throw", record.arg = exception, context.next = loc, caught && (context.method = "next", context.arg = undefined), !!caught; } for (var i = this.tryEntries.length - 1; i >= 0; --i) { var entry = this.tryEntries[i], record = entry.completion; if ("root" === entry.tryLoc) return handle("end"); if (entry.tryLoc <= this.prev) { var hasCatch = hasOwn.call(entry, "catchLoc"), hasFinally = hasOwn.call(entry, "finallyLoc"); if (hasCatch && hasFinally) { if (this.prev < entry.catchLoc) return handle(entry.catchLoc, !0); if (this.prev < entry.finallyLoc) return handle(entry.finallyLoc); } else if (hasCatch) { if (this.prev < entry.catchLoc) return handle(entry.catchLoc, !0); } else { if (!hasFinally) throw new Error("try statement without catch or finally"); if (this.prev < entry.finallyLoc) return handle(entry.finallyLoc); } } } }, abrupt: function abrupt(type, arg) { for (var i = this.tryEntries.length - 1; i >= 0; --i) { var entry = this.tryEntries[i]; if (entry.tryLoc <= this.prev && hasOwn.call(entry, "finallyLoc") && this.prev < entry.finallyLoc) { var finallyEntry = entry; break; } } finallyEntry && ("break" === type || "continue" === type) && finallyEntry.tryLoc <= arg && arg <= finallyEntry.finallyLoc && (finallyEntry = null); var record = finallyEntry ? finallyEntry.completion : {}; return record.type = type, record.arg = arg, finallyEntry ? (this.method = "next", this.next = finallyEntry.finallyLoc, ContinueSentinel) : this.complete(record); }, complete: function complete(record, afterLoc) { if ("throw" === record.type) throw record.arg; return "break" === record.type || "continue" === record.type ? this.next = record.arg : "return" === record.type ? (this.rval = this.arg = record.arg, this.method = "return", this.next = "end") : "normal" === record.type && afterLoc && (this.next = afterLoc), ContinueSentinel; }, finish: function finish(finallyLoc) { for (var i = this.tryEntries.length - 1; i >= 0; --i) { var entry = this.tryEntries[i]; if (entry.finallyLoc === finallyLoc) return this.complete(entry.completion, entry.afterLoc), resetTryEntry(entry), ContinueSentinel; } }, catch: function _catch(tryLoc) { for (var i = this.tryEntries.length - 1; i >= 0; --i) { var entry = this.tryEntries[i]; if (entry.tryLoc === tryLoc) { var record = entry.completion; if ("throw" === record.type) { var thrown = record.arg; resetTryEntry(entry); } return thrown; } } throw new Error("illegal catch attempt"); }, delegateYield: function delegateYield(iterable, resultName, nextLoc) { return this.delegate = { iterator: values(iterable), resultName: resultName, nextLoc: nextLoc }, "next" === this.method && (this.arg = undefined), ContinueSentinel; } }, exports; }
function asyncGeneratorStep(gen, resolve, reject, _next, _throw, key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { Promise.resolve(value).then(_next, _throw); } }
function _asyncToGenerator(fn) { return function () { var self = this, args = arguments; return new Promise(function (resolve, reject) { var gen = fn.apply(self, args); function _next(value) { asyncGeneratorStep(gen, resolve, reject, _next, _throw, "next", value); } function _throw(err) { asyncGeneratorStep(gen, resolve, reject, _next, _throw, "throw", err); } _next(undefined); }); }; }
var $ = function $(id) {
  return document.getElementById(id);
};
var $$ = function $$(clazz) {
  return document.getElementsByClassName(clazz);
};
var ui = ui || {};
ui.country_name={"ascension-island":"ac","andorra":"ad","united-arab-emirates":"ae","afghanistan":"af","antigua-and-barbuda":"ag","anguilla":"ai","albania":"al","armenia":"am","angola":"ao","antarctica":"aq","argentina":"ar","american-samoa":"as","austria":"at","australia":"au","aruba":"aw","aland-islands":"ax","azerbaijan":"az","bosnia-and-herzegovina":"ba","barbados":"bb","bangladesh":"bd","belgium":"be","burkina-faso":"bf","bulgaria":"bg","bahrain":"bh","burundi":"bi","benin":"bj","saint-barthélemy":"bl","bermuda":"bm","brunei-darussalam":"bn","bolivia":"bo","bonaire,-sint-eustatius-and-saba":"bq","brazil":"br","bahamas":"bs","bhutan":"bt","bouvet-island":"bv","botswana":"bw","belarus":"by","belize":"bz","canada":"ca","cocos-(keeling)-islands":"cc","democratic-republic-of-the-congo":"cd","central-european-free-trade-agreement":"cefta","central-african-republic":"cf","republic-of-the-congo":"cg","switzerland":"ch","côte-d'ivoire":"ci","cook-islands":"ck","chile":"cl","cameroon":"cm","china":"cn","colombia":"co","clipperton-island":"cp","costa-rica":"cr","cuba":"cu","cabo-verde":"cv","curaçao":"cw","christmas-island":"cx","cyprus":"cy","czech-republic":"cz","germany":"de","diego-garcia":"dg","djibouti":"dj","denmark":"dk","dominica":"dm","dominican-republic":"do","algeria":"dz","ceuta-&-melilla":"ea","ecuador":"ec","estonia":"ee","egypt":"eg","western-sahara":"eh","eritrea":"er","spain":"es","catalonia":"es-ct","galicia":"es-ga","ethiopia":"et","europe":"eu","finland":"fi","fiji":"fj","falkland-islands":"fk","federated-states-of-micronesia":"fm","faroe-islands":"fo","france":"fr","gabon":"ga","united-kingdom":"gb","england":"gb-eng","northern-ireland":"gb-nir","scotland":"gb-sct","wales":"gb-wls","grenada":"gd","georgia":"ge","french-guiana":"gf","guernsey":"gg","ghana":"gh","gibraltar":"gi","greenland":"gl","gambia":"gm","guinea":"gn","guadeloupe":"gp","equatorial-guinea":"gq","greece":"gr","south-georgia-and-the-south-sandwich-islands":"gs","guatemala":"gt","guam":"gu","guinea-bissau":"gw","guyana":"gy","hong-kong":"hk","heard-island-and-mcdonald-islands":"hm","honduras":"hn","croatia":"hr","haiti":"ht","hungary":"hu","canary-islands":"ic","indonesia":"id","ireland":"ie","israel":"il","isle-of-man":"im","india":"in","british-indian-ocean-territory":"io","iraq":"iq","iran":"ir","iceland":"is","italy":"it","jersey":"je","jamaica":"jm","jordan":"jo","japan":"jp","kenya":"ke","kyrgyzstan":"kg","cambodia":"kh","kiribati":"ki","comoros":"km","saint-kitts-and-nevis":"kn","north-korea":"kp","south-korea":"kr","kuwait":"kw","cayman-islands":"ky","kazakhstan":"kz","laos":"la","lebanon":"lb","saint-lucia":"lc","liechtenstein":"li","sri-lanka":"lk","liberia":"lr","lesotho":"ls","lithuania":"lt","luxembourg":"lu","latvia":"lv","libya":"ly","morocco":"ma","monaco":"mc","moldova":"md","montenegro":"me","saint-martin":"mf","madagascar":"mg","marshall-islands":"mh","north-macedonia":"mk","mali":"ml","myanmar":"mm","mongolia":"mn","macau":"mo","northern-mariana-islands":"mp","martinique":"mq","mauritania":"mr","montserrat":"ms","malta":"mt","mauritius":"mu","maldives":"mv","malawi":"mw","mexico":"mx","malaysia":"my","mozambique":"mz","namibia":"na","new-caledonia":"nc","niger":"ne","norfolk-island":"nf","nigeria":"ng","nicaragua":"ni","netherlands":"nl","norway":"no","nepal":"np","nauru":"nr","niue":"nu","new-zealand":"nz","oman":"om","panama":"pa","peru":"pe","french-polynesia":"pf","papua-new-guinea":"pg","philippines":"ph","pakistan":"pk","poland":"pl","saint-pierre-and-miquelon":"pm","pitcairn":"pn","puerto-rico":"pr","state-of-palestine":"ps","portugal":"pt","palau":"pw","paraguay":"py","qatar":"qa","réunion":"re","romania":"ro","serbia":"rs","russia":"ru","rwanda":"rw","saudi-arabia":"sa","solomon-islands":"sb","seychelles":"sc","sudan":"sd","sweden":"se","singapore":"sg","saint-helena,-ascension-and-tristan-da-cunha":"sh","slovenia":"si","svalbard-and-jan-mayen":"sj","slovakia":"sk","sierra-leone":"sl","san-marino":"sm","senegal":"sn","somalia":"so","suriname":"sr","south-sudan":"ss","sao-tome-and-principe":"st","el-salvador":"sv","sint-maarten":"sx","syria":"sy","eswatini":"sz","tristan-da-cunha":"ta","turks-and-caicos-islands":"tc","chad":"td","french-southern-territories":"tf","togo":"tg","thailand":"th","tajikistan":"tj","tokelau":"tk","timor-leste":"tl","turkmenistan":"tm","tunisia":"tn","tonga":"to","turkey":"tr","trinidad-and-tobago":"tt","tuvalu":"tv","taiwan":"tw","tanzania":"tz","ukraine":"ua","uganda":"ug","united-states-minor-outlying-islands":"um","united-nations":"un","united-states-of-america":"us","uruguay":"uy","uzbekistan":"uz","holy-see":"va","saint-vincent-and-the-grenadines":"vc","venezuela":"ve","virgin-islands-(british)":"vg","virgin-islands-(u.s.)":"vi","vietnam":"vn","vanuatu":"vu","wallis-and-futuna":"wf","samoa":"ws","kosovo":"xk","unknown":"xx","yemen":"ye","mayotte":"yt","south-africa":"za","zambia":"zm","zimbabwe":"zw"};
ui.country_name["united-states"] = "us";
ui.country_name["czechia"] = "cz";

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
  insertRowToOrderPos($("peer_list"), 2, row);
}
function cmpTime(a, b) {
  a = a.textContent.trim() || "999999";
  b = b.textContent.trim() || "999999";
  return a.localeCompare(b, 'en', {
    numeric: true
  });
}
function insertRowToOrderPos(tb, col, row) {
  var tr = tb.rows;
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
        'Riv-Save-Config': 'true'
      },
      body: JSON.stringify(peer_list.map(function (x) {
        return {
          "url": x
        };
      }))
    }).catch(function (error) {
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
  if (val.toFixed() * 1 > 99) {
    i += 1;
    val /= 1024;
  } else if (val.toFixed() * 1 > 9) {
    fixed = 1;
  }
  return val.toFixed(fixed) + ' ' + ['B/s', 'kB/s', 'MB/s', 'GB/s', 'TB/s'][i];
}
ui.showAllPeers = /*#__PURE__*/_asyncToGenerator( /*#__PURE__*/_regeneratorRuntime().mark(function _callee() {
  var response, peerList, peers;
  return _regeneratorRuntime().wrap(function _callee$(_context) {
    while (1) switch (_context.prev = _context.next) {
      case 0:
        _context.prev = 0;
        _context.next = 3;
        return fetch('api/publicpeers');
      case 3:
        response = _context.sent;
        _context.next = 6;
        return response.json();
      case 6:
        peerList = _context.sent;
        showWindow();
        ui.peers_country = Object.keys(peerList).flatMap(function (country) {
          return Object.keys(peerList[country]).map(function (peer) {
            var r = {};
            r[peer] = country.replace(".md", "");
            return r;
          });
        });
        ui.peers_country = ui.peers_country.reduce(function (r, c) {
          return Object.assign(r, c);
        }, {});
        peers = Object.values(peerList).flatMap(function (x) {
          return Object.keys(x);
        }); //start peers test
        _context.next = 13;
        return fetch('api/health', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(peers)
        });
      case 13:
        _context.next = 18;
        break;
      case 15:
        _context.prev = 15;
        _context.t0 = _context["catch"](0);
        console.error('Error:', _context.t0);
      case 18:
      case "end":
        return _context.stop();
    }
  }, _callee, null, [[0, 15]]);
}));
ui.getConnectedPeers = function () {
  return fetch('api/peers').then(function (response) {
    return response.json();
  });
};
ui.updateConnectedPeersHandler = function (peers) {
  ui.updateStatus(peers);
  $("peers").innerText = "";
  ui.connectedPeersAddress = peers.map(function (peer) {
    return peer.remote;
  });
  if (peers) {
    var regexStrip = /%[^\]]*/gm;
    peers.forEach(function (peer) {
      var row = $("peers").appendChild(document.createElement('div'));
      row.className = "overflow-ellipsis";
      var flag = row.appendChild(document.createElement("span"));
      if (peer.multicast || !peer.country_short) flag.className = "fas fa-thin fa-share-nodes peer-connected-fl";else flag.className = "fi fi-" + peer.country_short.toLowerCase() + " peer-connected-fl";
      row.append(peer.remote.replace(regexStrip, ""));
    });
  }
};
ui.updateStatus = function (peers) {
  var status = "st-error";
  if (peers) {
    if (peers.length) {
      var isNonMulticastExists = peers.filter(function (peer) {
        return !peer.multicast;
      }).length;
      status = isNonMulticastExists ? "st-multicast" : "st-connected";
    } else {
      status = "st-connecting";
    }
  }
  Array.from($$("status")).forEach(function (node) {
    return node.classList.add("is-hidden");
  });
  $(status).classList.remove("is-hidden");
};
ui.updateSpeed = function (peers) {
  if (peers) {
    var rsbytes = {
      "bytes_recvd": peers.reduce(function (acc, peer) {
        return acc + peer.bytes_recvd;
      }, 0),
      "bytes_sent": peers.reduce(function (acc, peer) {
        return acc + peer.bytes_sent;
      }, 0),
      "timestamp": Date.now()
    };
    if ("_rsbytes" in ui) {
      $("dn_speed").innerText = humanReadableSpeed((rsbytes.bytes_recvd - ui._rsbytes.bytes_recvd) * 1000 / (rsbytes.timestamp - ui._rsbytes.timestamp));
      $("up_speed").innerText = humanReadableSpeed((rsbytes.bytes_sent - ui._rsbytes.bytes_sent) * 1000 / (rsbytes.timestamp - ui._rsbytes.timestamp));
    }
    ui._rsbytes = rsbytes;
  } else {
    delete ui._rsbytes;
    $("dn_speed").innerText = humanReadableSpeed(-1);
    $("up_speed").innerText = humanReadableSpeed(-1);
  }
};
ui.updateConnectedPeers = function () {
  return ui.getConnectedPeers().then(function (peers) {
    return ui.updateConnectedPeersHandler(peers);
  }).catch(function (error) {
    ui.updateConnectedPeersHandler();
    $("peers").innerText = error.message;
  });
};
ui.getSelfInfo = function () {
  return fetch('api/self').then(function (response) {
    return response.json();
  });
};
ui.updateSelfInfo = function () {
  return ui.getSelfInfo().then(function (info) {
    $("ipv6").innerText = info.address;
    $("subnet").innerText = info.subnet;
    $("coordinates").innerText = ''.concat('[', info.coords.join(' '), ']');
    $("pub_key").innerText = info.key;
    $("priv_key").innerText = info.private_key;
    $("ipv6").innerText = info.address;
    $("version").innerText = info.build_version;
  }).catch(function (error) {
    $("ipv6").innerText = error.message;
  });
};
function main() {
  window.addEventListener("load", function () {
    $("showAllPeersBtn").addEventListener("click", ui.showAllPeers);
    ui.updateConnectedPeers();
    ui.updateSelfInfo();
    ui.sse = new EventSource('api/sse');
    ui.sse.addEventListener("health", function (e) {
      setHealth(JSON.parse(e.data));
    });
    ui.sse.addEventListener("peers", function (e) {
      ui.updateConnectedPeersHandler(JSON.parse(e.data));
    });
    ui.sse.addEventListener("rxtx", function (e) {
      ui.updateSpeed(JSON.parse(e.data));
    });
    ui.sse.addEventListener("coord", function (e) {
      var coords = JSON.parse(e.data);
      $("coordinates").innerText = ''.concat('[', coords.join(' '), ']');
    });
  });
}
main();
