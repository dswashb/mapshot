-- Automatically generated, do not modify
local data = {}
data.version = "0.0.7"
data.version_hash = "942832a35c1bc40fc7df7174704f2a7b02f2058c4f9b4778efa6a54b675e4389"
data.files = {}
data.files["index.html"] = function() return [==[
<html><head><title>Mapshot</title><style>body,html{margin:0}</style><link rel="stylesheet" href="https://unpkg.com/leaflet@1.7.1/dist/leaflet.css" integrity="sha512-xodZBNTC5n17Xt2atTPuE1HxjVMSvLVW9ocqUKLsCC5CXdbqCmblAshOMAS6/keqq/sMZMZ19scR4PsZChSR7A==" crossorigin=""><script src="https://unpkg.com/leaflet@1.7.1/dist/leaflet.js" integrity="sha512-XQoYMqMTK8LvdxXYG3nZ448hOEQiglfqkJs1NOQV44cWnUrBc8PkAOcXy20w0vlaXaVUearIOBhiXZ5V3ynxwA==" crossorigin=""></script><script>let MAPSHOT_CONFIG={};try{MAPSHOT_CONFIG=__MAPSHOT_CONFIG_TOKEN__}catch(_){}</script></head><body><div id="map" style="height:100%"></div><script src="./main-1c3f7217.js" defer=""></script></body></html>]==] end

data.files["main-1c3f7217.js"] = function() return [==[
(function () {
    'use strict';

    var _a, _b;
    const params = new URLSearchParams(window.location.search);
    let path = (_b = (_a = params.get("path")) !== null && _a !== void 0 ? _a : MAPSHOT_CONFIG.path) !== null && _b !== void 0 ? _b : "";
    if (!!path && path[path.length - 1] != "/") {
        path = path + "/";
    }
    console.log("Path", path);
    fetch(path + 'mapshot.json')
        .then(resp => resp.json())
        .then((info) => {
        console.log("Map info", info);
        const isIterable = function (obj) {
            // falsy value is javascript includes empty string, which is iterable,
            // so we cannot just check if the value is truthy.
            if (obj === null || obj === undefined) {
                return false;
            }
            return typeof obj[Symbol.iterator] === "function";
        };
        const worldToLatLng = function (x, y) {
            const ratio = info.render_size / info.tile_size;
            return L.latLng(-y * ratio, x * ratio);
        };
        const midPointToLatLng = function (bbox) {
            return worldToLatLng((bbox.left_top.x + bbox.right_bottom.x) / 2, (bbox.left_top.y + bbox.right_bottom.y) / 2);
        };
        const baseLayer = L.tileLayer(path + "zoom_{z}/tile_{x}_{y}.jpg", {
            tileSize: info.render_size,
            bounds: L.latLngBounds(worldToLatLng(info.world_min.x, info.world_min.y), worldToLatLng(info.world_max.x, info.world_max.y)),
            noWrap: true,
            maxNativeZoom: info.zoom_max,
            minNativeZoom: info.zoom_min,
            minZoom: info.zoom_min - 4,
            maxZoom: info.zoom_max + 4,
        });
        const debugLayers = [
            L.marker([0, 0], { title: "Start" }).bindPopup("Starting point"),
        ];
        if (info.player) {
            debugLayers.push(L.marker(worldToLatLng(info.player.x, info.player.y), { title: "Player" }).bindPopup("Player"));
        }
        debugLayers.push(L.marker(worldToLatLng(info.world_min.x, info.world_min.y), { title: `${info.world_min.x}, ${info.world_min.y}` }), L.marker(worldToLatLng(info.world_min.x, info.world_max.y), { title: `${info.world_min.x}, ${info.world_max.y}` }), L.marker(worldToLatLng(info.world_max.x, info.world_min.y), { title: `${info.world_max.x}, ${info.world_min.y}` }), L.marker(worldToLatLng(info.world_max.x, info.world_max.y), { title: `${info.world_max.x}, ${info.world_max.y}` }));
        let stationsLayers = [];
        if (isIterable(info.stations)) {
            for (const station of info.stations) {
                stationsLayers.push(L.marker(midPointToLatLng(station.bounding_box), { title: station.backer_name }).bindTooltip(station.backer_name, { permanent: true }));
            }
        }
        let tagsLayers = [];
        if (isIterable(info.tags)) {
            for (const tag of info.tags) {
                tagsLayers.push(L.marker(worldToLatLng(tag.position.x, tag.position.y), { title: `${tag.force_name}: ${tag.text}` }).bindTooltip(tag.text, { permanent: true }));
            }
        }
        const mymap = L.map('map', {
            crs: L.CRS.Simple,
            layers: [baseLayer],
        });
        L.control.layers({ /* Only one default base layer */}, {
            "Train stations": L.layerGroup(stationsLayers),
            "Tags": L.layerGroup(tagsLayers),
            "Debug": L.layerGroup(debugLayers),
        }).addTo(mymap);
        mymap.setView([0, 0], 0);
    });

}());
//# sourceMappingURL=main-1c3f7217.js.map
]==] end

return data
