"use strict";
  
function initMap() {
    const map = new google.maps.Map(document.getElementById('map'), {
        zoom: 2,
        center: new google.maps.LatLng(2.8,-187.3),
        mapTypeControl: false,
    });

    const infoWindow = new google.maps.InfoWindow();
    infoWindow.setOptions({pixelOffset: new google.maps.Size(0, -30)});

    // Fit to markers when everything has settled down.
    const bounds = new google.maps.LatLngBounds();
    google.maps.event.addListenerOnce(map, 'idle', () => {
        map.fitBounds(bounds);
        searchBox.setBounds(bounds);
    });

    // Extend bounds whenever a feature is added.
    map.data.addListener('addfeature', event => {
        bounds.extend(event.feature.getGeometry().get());
    });

    // Create the search box and link it to the UI element.
    var input = document.getElementById('pac-input');
    var searchBox = new google.maps.places.SearchBox(input);
    map.controls[google.maps.ControlPosition.TOP_LEFT].push(input);

    searchBox.addListener('places_changed', () => {
        let places = searchBox.getPlaces();
        if (places.length == 0) {
            return;
        }
        fitToNearestMarkers(places[0].geometry.location, map);
    });

    // Load data.
    map.data.loadGeoJson('/features');

    map.data.setStyle(feature => {
        return {
            icon: {
                url: "/static/img/spotlight-poi-pink.png"
            }
        };
    });

    // Show the information for a class when its marker is clicked.
    map.data.addListener('click', event => {
        const title = event.feature.getProperty('title');
        const address = event.feature.getProperty('address');
        const locationUrl = event.feature.getProperty('locationUrl');
        const position = event.feature.getGeometry().get();
        const content = sanitizeHTML`
            <div id="content">
              <h4>${title}</h4>
              <div id="bodyContent">
                <p>${address}</p>
                <p>Link: <a href="${locationUrl}" target="_blank">${locationUrl}</a></p>
              </div>
            </div>
        `;

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
  const entities = {'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'};
  let result = strings[0];
  for (let i = 1; i < arguments.length; i++) {
    result += String(arguments[i]).replace(/[&<>'"]/g, (char) => {
      return entities[char];
    });
    result += strings[i];
  }
  return result;
}

function fitToNearestMarkers(location, map) {
    let distances = [];
    map.data.forEach(feature => {
        let markerLocation = feature.getGeometry().get();
        let dist = google.maps.geometry.spherical.computeDistanceBetween(location, markerLocation);
        distances.push([dist, markerLocation, feature]);
    });

    let nearest = distances.sort((a, b) => {
        if (a[0] < b[0]) {return -1;}
        if (a[0] > b[0]) {return 1;}
        return 0;
    }).slice(0, 5);

    let bounds = new google.maps.LatLngBounds();
    bounds.extend(location); // Make sure to include the searched for location.
    nearest.forEach(dist => {
        bounds.extend(dist[1]);
    });
    map.fitBounds(bounds);
}