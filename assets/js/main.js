
//渲染器字典
var Renderer = {};

//运行函数字典
var Runner = {};

var Edit = {};

var Remove = {};

var Stop = {};

var Save = {}

//Modal
var Modal = {};

//图标字典
var Icons = {
    "add": `<svg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 512 512'>
                <line x1='256' y1='112' x2='256' y2='400'
                    style='fill:none;stroke:#000;stroke-linecap:round;stroke-linejoin:round;stroke-width:32px' />
                <line x1='400' y1='256' x2='112' y2='256'
                    style='fill:none;stroke:#000;stroke-linecap:round;stroke-linejoin:round;stroke-width:32px' />
            </svg>`,
    "run": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
                <title>start</title>
                <path d="M8 5v14l11-7z" />
                <path d="M0 0h24v24H0z" fill="none" />
            </svg>`,
    "stop": `<svg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 512 512'>
                <title>stop</title>
                <rect x='96' y='96' width='320' height='320' rx='24' ry='24' 
                    style='fill:none;stroke:#000;stroke-linejoin:round;stroke-width:32px'/>
            </svg>`,
    "remove": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
                <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM8 9h8v10H8V9zm7.5-5l-1-1h-5l-1 1H5v2h14V4z" />
                <path fill="none" d="M0 0h24v24H0V0z" />
                </svg>`,
    "edit": `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
                <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
                <path d="M0 0h24v24H0z" fill="none" />
            </svg>`
};

//弹出错误提示
function NewError(message) {
    var errorMessage = format(`<div class="alert alert-warning">
                <button type="button" class="close" data-dismiss="alert">&times;</button>
                {0}
              </div>`, [message]);
    document.getElementById("flashMessage").innerHTML = errorMessage;
}

//格式化字符串
function format(source, params) {
    $.each(params, function (i, n) {
        source = source.replace(new RegExp("\\{" + i + "\\}", "g"), n);
    })
    return source;
}

//编码base64
function utoa(str) {
    unicode = window.btoa(unescape(encodeURIComponent(str)));
    unicode = unicode.replace(/\+/g, "-");
    unicode = unicode.replace(/\//g, '_');
    unicode = unicode.replace(/\=/g, '');
    return unicode;
}

//解码base64
function atou(str) {
    base64 = decodeURIComponent(str);
    base64 = base64.replace(/\-/g, "+");
    base64 = base64.replace(/\_/g, '-');
    i = base64.length % 4
    if (i != 0) {
        base64 += "=".repeat(4 - i)
    }
    console.log(base64)
    result = decodeURIComponent(escape(window.atob(base64)))
    console.log(result)
    return result
}

function RegisterStop() {
    Stop["formatted"] = stopFormatted;
}
function stopFormatted(data = []) {
    $.post("/api/status", JSON.stringify({ "cmd": "del", "id": { "id": data["id"], "type": data["type"] } })).fail(function (request) {
        NewError(request.responseJSON["error"])
    }).fail(function (request) {
        NewError(request.responseJSON["error"])
    });
}
function RegisterRemove() {
    Remove["formatted"] = removeFormatted;
    Remove["sub"] = removeSub;
}
function removeFormatted(data = []) {
    $.post(format("/api/configs/{0}/{1}", [data["type"], data["id"]]), JSON.stringify({ "cmd": "del" })).fail(function (request) {
        NewError(request.responseJSON["error"])
    }).fail(function (request) {
        NewError(request.responseJSON["error"])
    });
}
function removeSub(data = []) {
    $.post(format("/api/sub/{0}", [data["id"]]), JSON.stringify({ "cmd": "del" })).fail(function (request) {
        NewError(request.responseJSON["error"])
    }).fail(function (request) {
        NewError(request.responseJSON["error"])
    });
}
function RegisterSave() {
    Save["formatted"] = saveFormatted;
    Save["sub"] = saveSub;
}
function saveFormatted(data = {}) {
    data["id"] = (data["id"] == undefined) ? "" : data["id"];
    $.post(format("/api/configs/formatted/{0}", [data["id"]]),
        JSON.stringify({
            "cmd": "edit",
            "id": data["id"],
            'name': document.getElementById('name').value,
            'group': document.getElementById('group').value,
            'port': document.getElementById('port').value,
            'boundType': document.getElementById('boundType').value,
            'protocolType': document.getElementById('protocolType').value,
            'protocolSettings': JSON.parse(document.getElementById('protocolSettings').value == "" ? null : document.getElementById('protocolSettings').value),
            'mux': document.getElementById('mux').value,
            'streamSettings': JSON.parse(document.getElementById('streamSettings').value == "" ? null : document.getElementById('streamSettings').value),
            'proxyID': document.getElementById('proxyID').value == "" ? null : document.getElementById('proxyID').value
        })).fail(function (request) {
            NewError(request.responseJSON["error"]);
        });
}
function saveSub(data = []) {
    data["id"] = (data["id"] == undefined) ? "" : data["id"];
    $.post(format("/api/sub/{0}", [data["id"]]),
        JSON.stringify({
            "cmd": "edit",
            "id": data["id"],
            "name": document.getElementById('name').value,
            "url": document.getElementById('url').value
        })).fail(function (request) {
            NewError(request.responseJSON["error"])
        })
}
function RegisterEdit() {
    Edit["formatted"] = editFormatted;
    Edit["sub"] = editSubscription;
}
function editFormatted(data) {
    $.getJSON(format("/api/configs/{0}/{1}", [data["type"], data["id"]]), Modal["formatted"]
    ).fail(function (request) {
        NewError(request.responseJSON["error"])
    });
}

function editSubscription(data) {
    $.getJSON(format("/api/sub/{1}", [data["type"], data["id"]]), Modal["sub"]).fail(function (request) {
        NewError(request.responseJSON["error"])
    }).fail(function (request) {
        NewError(request.responseJSON["error"])
    });
}
//注册渲染器
function RegisterRederer() {
    Renderer['list'] = rendererList;
    Renderer['lists'] = rendererLists;
    Renderer['grid'] = rendererGrid;
    Renderer['input'] = rendererInput;
    Renderer['textare'] = rendererTextarea;
    Renderer['select'] = rendererSelect;
    Renderer['hr'] = rendererHR;
    Renderer['button'] = rendererButton;
    Renderer['hyperlink'] = rendererHyperlink;
}

function rendererHR(data = [], event = []) {
    `
    {
        "label":'label'
    }
    `
    var result = format(`
    <hr>
    <div class="alert alert-light" role="alert">
            {0}
    </div>
    `, [data["label"]]);
    return [result, [], []];
}
function rendererInput(data = [], event = []) {
    `{
        "label":"label",
        "id":"id",
        "value":""
    }`
    var value = "";
    if (data["value"] != undefined) {
        value = data["value"];
    }
    var id = utoa(JSON.stringify({ "label": data["label"], "renderer": "input" }));
    if (data["id"] != undefined) {
        id = data["id"];
    }
    return [format(`
    <div class="form-group">
        <label for="{1}">{0}</label>
        <input type="input" class="form-control" id="{1}" value="{2}">
    </div>
    `, [data["label"], id, value]), [id], []]
}

function rendererSelect(data = [], event = []) {
    `
    {
        "label":"label",
        "selected":["option","value"],
        "option":[["option","value"]],
        "id":"id"
    }`;
    var id = utoa(JSON.stringify({ "label": data["label"], "renderer": "select" }));
    if (data["id"] != undefined) {
        id = data["id"];
    }
    var option = "";
    if (data['selected'] != undefined) {
        var option = format("<option selected value={0}>{1}</option>", [data['selected'][1], data['selected'][0]]);
    }
    for (var key in data["option"]) {
        option += format("<option value={0}>{1}</option>", [data["option"][key][1], data["option"][key][0]]);
    }
    var result = format(`
    <div class="form-group">
        <label for="{0}">{1}</label>
        <select id="{0}" class="form-control">
            {2}
        </select>
    </div>
    `, [id, data["label"], option]);
    return [result, [id], []]
}

function rendererTextarea(data = [], event = []) {
    `
    {
        "label":"label",
        "row":"row",
        "id":"id",
        "value":"value"
    }
    `;
    var value = "";
    if (data["value"] != undefined) {
        value = data["value"];
    }
    var id = utoa(JSON.stringify({ "label": data["label"], "renderer": "input" }));
    if (data["id"] != undefined) {
        id = data["id"];
    }
    return [format(`
    <div class="form-group">
        <label for="{0}">{1}</label>
        <textarea class="form-control" id="{0}" rows="{2}">{3}</textarea>
    </div>
    `, [id, data["label"], data["row"], value]), [id], []]
}

function rendererGrid(data = [], event = []) {
    data["data"] = Renderer[data["data"][0]](data["data"][1], event);
    return renderergrid(data);
}
//渲染
function renderergrid(data = [], event = []) {
    `[
        "row":[1,2,3,4],
        "data":[
            [data,[id],[event]]
        ],
    ]`
    var id = [];
    var event = [];
    var result = format(`
    <div class="container">
        <div class="row row-cols-{0} row-cols-sm-{1} row-cols-md-{2} row-cols-md-{3}">
    `, [data['row'][0], data['row'][1], data['row'][2], data['row'][3]]);
    for (var key in data["data"]) {
        result += data["data"][key][0];
        for (var eventKey in data["data"][key][2]) {
            event.push(data["data"][key][2][eventKey]);
        }
        for (var idKey in data["data"][key][1]) {
            id.push(data["data"][key][1][idKey]);
        }
    }
    result += `
    </div>
    </div>`
    return [result, id, event]
}

function rendererLists(data = [], event = []) {
    if (data == undefined) {
        return [["", [], []]];
    }
    var result = []
    for (var key in data) {
        result.push(rendererList(data[key], event));
    }
    return result;
}

function rendererList(data = [], event = []) {
    `
    {
        "label":"label",
        "subtitle":"subtitle",
        "icon":["1","2","3"],
        "id":["1","2","3"]
    }
    `;
    `
    [
        [[事件,函數,參數]],
    ]`
    if (data == undefined) {
        return ["", [], []];
    };
    var id1 = utoa(JSON.stringify({ "label": data["label"], "renderer": "select" }));
    var id2 = utoa(JSON.stringify({ "label": data["label"], "renderer": "select" }));
    var id3 = utoa(JSON.stringify({ "label": data["label"], "renderer": "select" }));
    if (data["id"] != undefined) {
        id1 = data["id"][0];
        id2 = data["id"][1];
        id3 = data["id"][2];
    }
    listString = format(`
    <div class="col-sm rounded border shadow bg-white custom-control custom-radio" style="min-height: 100px;">
        <label class="w-100 h-100">
            <label class="custom-control-label" >{1}</label>      
        <div class="alert alert-light" role="alert">
        {0}
        </div>
            <a id="{5}" class="float-right" style="z-index:1;">
                {2}
            </a>
            <a id="{6}" class="float-right" style="z-index:1;">
                {3}
            </a>
            <a id="{7}" class="float-right"  style="z-index:1;">
                {4}
            </a>
        
        </label>
    </div>`, [data["subtitle"], data["label"], Icons[data["icon"][0]], Icons[data["icon"][1]], Icons[data["icon"][2]], id1, id2, id3]);
    var revent = [];
    for (var key in event[0]) {
        revent.push([id1, event[0][key][0], event[0][key][1], event[0][key][2]]);
    }
    for (var key in event[1]) {
        revent.push([id2, event[1][key][0], event[1][key][1], event[2][key][2]]);
    }
    for (var key in event[2]) {
        revent.push([id3, event[2][key][0], event[2][key][1], event[2][key][2]]);
    }
    return [listString, [id1, id2, id3], revent];
}
function rendererButton(data = [], event = []) {
    `
    {
        "label":"name",
        "class":"btn",
        "id":"",
        "data-mismiss":""
    }
    `
    var id = utoa(JSON.stringify({ "label": data["label"], "renderer": "button" }));
    if (data["id"] != undefined) {
        id = data["id"];
    }
    var result = format(`
    <button type="button" class="btn {0}" data-dismiss="{1}" id="{2}">{3}</button>`, [data["class"], data["data-dismiss"], id, data["label"]]);
    var revent = [];
    for (var key in event[0]) {
        revent.push([id, event[0][key][0], event[0][key][1], event[0][key][2]]);
    }
    return [result, [id], revent];

}
function rendererHyperlink(data = [], event = []) {
    `{
        "label":"label",
        "herf":"",
        "id":"",
    }`;
    var id = (data["id"] != undefined) ? data["id"] : "";

    var herf = (data["herf"] != undefined) ? data["herf"] : "";
    var result = format(`
    <a herf="{0}",id="{1}">{2}</a>`, [herf, id, data["label"]]);
    var revent = [];
    for (var key in event[0]) {
        revent.push([id, event[0][key][0], event[0][key][1], event[0][key][2]]);
    }
    return [result, [id], revent];

}
//注册运行函数
function RegisterRunner() {
    Runner["json"] = runnerConfig;
    Runner["formatted"] = runnerConfig;
    Runner["sub"] = runnerSub;
}

function runnerConfig(jsonData) {
    $.post(format("/api/status", []),
        JSON.stringify({ "cmd": "add", "id": { "id": jsonData['id'].toString(10), "type": jsonData['type'] } }),
        function (resulte, status) {
            if (status == "success") {
                refresh()
            } else {
                NewError(resulte["error"]);
            }
        }, "json").fail(function (request) {
            NewError(request.responseJSON["error"])
        });
}

function runnerSub(jsonData) {
    $.post(format("/api/sub/{0}", [jsonData["id"]]),
        JSON.stringify({ "cmd": "update" }),
        function (resulte, status) {
            if (status == "success") {
                refresh();
            } else {
                NewError(resulte["error"]);
            }
        }, "json").fail(function (request) {
            NewError(request.responseJSON["error"])
        });
}
//注册Model
function RegisterModal() {
    Modal["formatted"] = ModalFormatted;
    Modal["managesub"] = ModalManageSub;
    Modal["sub"] = ModalSubscription;
}

function ModalFormatted(data = []) {
    $.getJSON("/api/configs", function (result) {
        var proxyID = (data["proxyID"] != undefined && data["proxyID"] != "") ? data["proxyID"] : "";
        var proxyIDSelected=["",""]
        var proxyIDList =[["",""]]
        for (var key in result){
            if (result[key]["id"]["type"]=="formatted"&&result[key]["boundType"]=="outbound"){
                if (data["id"]==result[key]["id"]["id"]){
                    continue;
                }
                if (proxyID==result[key]["id"]["id"]){
                    proxyIDSelected=[result[key]["name"],result[key]["id"]["id"]];
                    continue
                }
                proxyIDList.push([result[key]["name"],result[key]["id"]["id"]]);
            }
        }
        var protocolSettings = (data["protocolSettings"] != undefined && data["protocolSettings"] != "") ? JSON.stringify(data["protocolSettings"], null, 4) : "";
        var streamSettings = (data["streamSettings"] != undefined && data["streamSettings"] != "") ? JSON.stringify(data["streamSettings"], null, 4) : "";
        
        var bodyHTML = [
            ["input", { "label": "name", "value": data["name"], "id": "name" }],
            ["input", { "label": "group", "value": (data["group"] != undefined) ? data["group"] : "default", "id": "group" }],
            ["input", { "label": "port", "value": (data["port"] != undefined) ? data["port"] : "0", "id": "port" }],
            ["select", {
                "label": "boundType",
                "id": "boundType",
                "selected": [data["boundType"], data["boundType"]],
                "option": [["inbound", "inbound"], ["outbound", "outbound"]]
            }],
            ["select", {
                "label": "protocolType",
                "id": "protocolType",
                "selected": [data["protocolType"], data["protocolType"]],
                "option": [
                    ["vmess", "vmess"],
                    ["shadowsocks", "shadowsocks"],
                    ["socks", "socks"],
                    ["http", "http"],
                    ["dokodemo-door","dokodemo-door"],
                    ["freedom","freedom"]],
            }],
            ["input", { "label": "mux", "value": (data["mux"] != undefined) ? data["mux"] : "0", "id": "mux" }],
            ["textare", {
                "label": "protocolSettings",
                "id": "protocolSettings",
                "row": "10",
                "value": protocolSettings
            }],
            ["textare", {
                "label": "streamSettings",
                "id": "streamSettings",
                "row": "10",
                "value": streamSettings
            }],
            ["select", {
                "label": "proxyID",
                "id": "proxyID",
                "selected": proxyIDSelected,
                "option": proxyIDList
            }],
        ]
        var header = ["添加配置", [], []];
        var footerHTML = [
            ["button", {
                "label": "關閉",
                "class": "btn btn-secondary",
                "id": "",
                "data-dismiss": "modal"
            }],
            ["button", {
                "label": "保存",
                "class": "btn btn-primary",
                "id": utoa(format(`{"id":"{0}","type":"{1}"}`,
                    [(data["id"] != undefined) ? data["id"] : "", "formatted"])),
                "data-dismiss": "modal"
            }, [[['click', save], ['click', refresh]]]],
        ]
        jump(header, RendererAll(bodyHTML), RendererAll(footerHTML))
    })
}

function ModalSubscription(data = []) {
    var bodyHTML = [
        ["input", { "label": "name:", "id": "name", "value": data["name"] }],
        ["textare", { "label": "URL:", "row": "3", "id": "url", "value": data["url"] }]
    ];
    var header = ["添加配置", [], []];
    var footerHTML = [
        ["button", {
            "label": "關閉",
            "class": "btn btn-secondary",
            "id": "",
            "data-dismiss": "modal"
        }],
        ["button", {
            "label": "保存",
            "class": "btn btn-primary",
            "id": utoa(format(`{"id":"{0}","type":"{1}"}`, [(data["id"] != undefined) ? data["id"] : "", "sub"])),
        }, [[['click', save], ['click', Modal["managesub"]]]]],
    ]
    jump(header, RendererAll(bodyHTML), RendererAll(footerHTML))
}
function ModalManageSub() {
    $.getJSON("/api/sub", function (result) {
        var subscription = [];
        for (var key in result) {
            var temp = {
                "label": result[key]["name"],
                "subtitle": result[key]["boundType"],
                "icon": ['run', 'edit', 'remove'],
                "id": [
                    utoa(format(`{"id":"{0}","type":"{1}","boundType":"{2}","number":"0"}`, [result[key]["id"]["id"], result[key]["id"]["type"], result[key]["boundType"]])),
                    utoa(format(`{"id":"{0}","type":"{1}","boundType":"{2}","number":"1"}`, [result[key]["id"]["id"], result[key]["id"]["type"], result[key]["boundType"]])),
                    utoa(format(`{"id":"{0}","type":"{1}","boundType":"{2}","number":"2"}`, [result[key]["id"]["id"], result[key]["id"]["type"], result[key]["boundType"]])),
                ]
            }
            subscription.push(temp);
        }
        var bodyHTML = [
            ["hr", { "label": "訂閲列表" }],
            ["grid", {
                "row": [1, 1, 1, 1],
                "data": [
                    "lists", subscription,
                ]
            }, [[['click', run]], [['click', edit]], [['click', remove]]]]];
        var header = ["訂閲設置", [], []];
        var footerHTML = [
            ["button", {
                "label": "關閉",
                "class": "btn btn-primary",
                "id": "save",
                "data-dismiss": "modal"
            }, [[['click', refresh]]]],
        ]
        jump(header, RendererAll(bodyHTML), RendererAll(footerHTML))
    }
    ).fail(function (request) {
        NewError(request.responseJSON["error"])
    })
}

//修改标签内容
function apply(id, data) {
    //data 为 [数据,id,绑定事件]数组
    //绑定事件为[[id,事件,函数,函数参数].....]数组
    $(format("#{0}", [id])).html(data[0]);
    $.each(data[2], function (key, item) {
        if (item[3] != undefined) {
            $(format("#{1}", [item[0]])).trigger(item[1], item[3]);
        }
        $(format("#{0}", [item[0]])).on(item[1], item[2]);
        // $(format("#{0}",[item[0]])).die().live(item[1],item[2])
    })
}

//刷新网页
function refresh() {
    index()
}

//主页初始化
function index() {
    $.getJSON("/api/status", function (status) {
        $.getJSON("/api/configs", function (result) {
            var data = { "running": [], "outbound": [], "inbound": [], "json": [] };
            for (var key in result) {
                var temp = {
                    "label": result[key]["name"],
                    "subtitle": result[key]["boundType"],
                    "icon": ['run', 'edit', 'remove'],
                    "id": [
                        utoa(format(`{"id":"{0}","type":"{1}","boundType":"{2}","number":"0"}`, [result[key]["id"]["id"], result[key]["id"]["type"], result[key]["boundType"]])),
                        utoa(format(`{"id":"{0}","type":"{1}","boundType":"{2}","number":"1"}`, [result[key]["id"]["id"], result[key]["id"]["type"], result[key]["boundType"]])),
                        utoa(format(`{"id":"{0}","type":"{1}","boundType":"{2}","number":"2"}`, [result[key]["id"]["id"], result[key]["id"]["type"], result[key]["boundType"]]))
                    ]
                }
                if (status['status'] && status['running'] &&
                    format("{'type':'{0}','id':'{1}'}", [result[key]['id']['type'],
                    result[key]['id']['id']]) in status['running']) {
                    temp["icon"] = ['stop', 'edit', 'remove']
                    data["running"].push(temp);
                } else if (result[key]["boundType"] == "outbound") {
                    data["outbound"].push(temp);
                } else if (result[key]["boundType"] == "inbound") {
                    data["inbound"].push(temp);
                } else if (result[key]["boundType"] == "json") {
                    data["json"].push(temp);
                }
            }
            var id = []
            bodyHTML = [
                ["hr", { "label": "正在运行" }],
                ["grid", {
                    "row": [1, 2, 3, 4],
                    "data": [
                        "lists", data["running"],
                    ]
                }, [[['click', stop]], [['click', edit]], [['click', remove]]]
                ],
                ["hr", { "label": "outbound" }],
                ["grid", {
                    "row": [1, 2, 3, 4],
                    "data": [
                        "lists", data["outbound"]
                    ]
                }, [[['click', run]], [['click', edit]], [['click', remove]]]
                ],
                ["hr", { "label": "inbound" }],
                ["grid", {
                    "row": [1, 2, 3, 4],
                    "data": [
                        "lists", data["inbound"]
                    ]
                }, [[['click', run]], [['click', edit]], [['click', remove]]]]

            ];
            addHTML = [
                ["button", {
                    "label": "管理訂閲",
                    "id": utoa(format(`{"type":"managesub"}`,
                        [])),
                }, [[["click", cmodal]]]],
                ["button", {
                    "label": "添加配置",
                    "id": utoa(format(`{"type":"formatted"}`,
                        [])),
                }, [[["click", cmodal]]]],
                ["button", {
                    "label": "添加訂閲",
                    "id": utoa(format(`{"type":"sub"}`,
                        [])),
                }, [[["click", cmodal]]]],
            ]
            var body = RendererAll(bodyHTML);
            apply('bodyList', body);
            apply('dropdownMenu', RendererAll(addHTML))
        }).fail(function (request) {
            NewError(request.responseJSON["error"])
        });
    });
}

function addObj(obja, objb) {
    obja[0] += objb[0];
    for (var key in objb[1]) {
        obja[1].push(objb[1][key]);
    }
    for (var key in objb[2]) {
        obja[2].push(objb[2][key]);
    }
    return obja;
}

//修改弹窗
function jump(header, body, footer) {
    apply('ModalScrollableTitle', header);
    apply('ModalScrollableBody', body);
    apply('ModalScrollableFooter', footer);
    $('#ModalScrollable').modal("show");
}

function RendererAll(data) {
    var body = ["", [], []]
    for (var key in data) {
        var temp = Renderer[data[key][0]](data[key][1], data[key][2]);
        body = addObj(body, temp);
    }
    return body;
}

//初始化网页
function init() {
    RegisterRederer();
    RegisterRunner();
    RegisterEdit()
    RegisterModal();
    RegisterSave();
    RegisterRemove();
    RegisterStop();
    index();
}

//运行
function run() {
    jsonData = JSON.parse(atou(event.currentTarget.id));
    Runner[jsonData["type"]](jsonData);
    refresh();
}

function remove() {
    jsonData = JSON.parse(atou(event.currentTarget.id));
    Remove[jsonData["type"]](jsonData);
    refresh();
}

function stop() {
    jsonData = JSON.parse(atou(event.currentTarget.id));
    Stop[jsonData["type"]](jsonData);
    refresh();

}

function save() {
    jsonData = JSON.parse(atou(event.currentTarget.id));
    Save[jsonData["type"]](jsonData);
    refresh();
}

function edit() {
    jsonData = JSON.parse(atou(event.currentTarget.id));
    Edit[jsonData["type"]](jsonData);
    refresh()
}
function cmodal() {
    jsonData = JSON.parse(atou(event.currentTarget.id));
    Modal[jsonData["type"]]();
    refresh();
}
//初始化网页
$(document).ready(init());
