"use strict";

var _templateObject = _taggedTemplateLiteral(['\n            <div id="content">\n              <h4>', '</h4>\n              <div id="bodyContent">\n                <p>', '</p>\n                <p>Link: <a href="', '" target="_blank">', '</a></p>\n              </div>\n            </div>\n        '], ['\n            <div id="content">\n              <h4>', '</h4>\n              <div id="bodyContent">\n                <p>', '</p>\n                <p>Link: <a href="', '" target="_blank">', '</a></p>\n              </div>\n            </div>\n        ']);

function _taggedTemplateLiteral(strings, raw) { return Object.freeze(Object.defineProperties(strings, { raw: { value: Object.freeze(raw) } })); }

function initMap() {
    var map = new google.maps.Map(document.getElementById('map'), {
        zoom: 2,
        center: new google.maps.LatLng(2.8, -187.3),
        mapTypeControl: false
    });

    var infoWindow = new google.maps.InfoWindow();
    infoWindow.setOptions({ pixelOffset: new google.maps.Size(0, -30) });

    // Fit to markers when everything has settled down.
    var bounds = new google.maps.LatLngBounds();
    google.maps.event.addListenerOnce(map, 'idle', function () {
        map.fitBounds(bounds);
        searchBox.setBounds(bounds);
    });

    // Extend bounds whenever a feature is added.
    map.data.addListener('addfeature', function (event) {
        bounds.extend(event.feature.getGeometry().get());
    });

    // Create the search box and link it to the UI element.
    var input = document.getElementById('pac-input');
    var searchBox = new google.maps.places.SearchBox(input);
    map.controls[google.maps.ControlPosition.TOP_LEFT].push(input);

    searchBox.addListener('places_changed', function () {
        var places = searchBox.getPlaces();
        if (places.length == 0) {
            return;
        }
        fitToNearestMarkers(places[0].geometry.location, map);
    });

    // Load data.
    map.data.loadGeoJson('/features');

    map.data.setStyle(function (feature) {
        return {
            icon: {
                url: "/static/img/spotlight-poi-pink.png"
            }
        };
    });

    // Show the information for a class when its marker is clicked.
    map.data.addListener('click', function (event) {
        var title = event.feature.getProperty('title');
        var address = event.feature.getProperty('address');
        var locationUrl = event.feature.getProperty('locationUrl');
        var position = event.feature.getGeometry().get();
        var content = sanitizeHTML(_templateObject, title, address, locationUrl, locationUrl);

        infoWindow.setContent(content);
        infoWindow.setPosition(position);
        infoWindow.open(map);

        map.panTo(position);
        if (map.getZoom() < 10) {
            map.setZoom(10);
        }
    });
}

// Escapes HTML characters in a template literal string, to prevent XSS.
function sanitizeHTML(strings) {
    var entities = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };
    var result = strings[0];
    for (var i = 1; i < arguments.length; i++) {
        result += String(arguments[i]).replace(/[&<>'"]/g, function (char) {
            return entities[char];
        });
        result += strings[i];
    }
    return result;
}

function fitToNearestMarkers(location, map) {
    var distances = [];
    map.data.forEach(function (feature) {
        var markerLocation = feature.getGeometry().get();
        var dist = google.maps.geometry.spherical.computeDistanceBetween(location, markerLocation);
        distances.push([dist, markerLocation, feature]);
    });

    var nearest = distances.sort(function (a, b) {
        if (a[0] < b[0]) {
            return -1;
        }
        if (a[0] > b[0]) {
            return 1;
        }
        return 0;
    }).slice(0, 5);

    var bounds = new google.maps.LatLngBounds();
    bounds.extend(location); // Make sure to include the searched for location.
    nearest.forEach(function (dist) {
        bounds.extend(dist[1]);
    });
    map.fitBounds(bounds);
}