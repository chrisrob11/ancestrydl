package commands

import (
	"fmt"
)

// generatePersonPageTemplate creates a single person page that uses URL parameters
func generatePersonPageTemplate(peopleJSON, metadataJSON string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Person - Family Tree</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            padding: 20px;
        }

        .container {
            max-width: 1000px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 30px;
        }

        .back-button {
            display: inline-block;
            padding: 10px 20px;
            background: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin-bottom: 20px;
        }

        .back-button:hover {
            background: #2980b9;
        }

        h1 {
            color: #2c3e50;
            margin-bottom: 5px;
            font-size: 2.5em;
        }

        .person-id {
            color: #95a5a6;
            font-size: 0.9em;
            margin-bottom: 20px;
        }

        .hero-section {
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 2px solid #e0e0e0;
        }

        .vital-stats {
            font-size: 1.2em;
            color: #555;
            margin: 10px 0;
        }

        .vital-stats .icon {
            margin-right: 8px;
        }

        .location-info {
            font-size: 1em;
            color: #666;
            margin: 8px 0;
        }

        .section {
            margin: 30px 0;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 4px;
        }

        .section h2 {
            color: #2c3e50;
            margin-bottom: 15px;
            font-size: 1.5em;
        }

        .info-grid {
            display: grid;
            grid-template-columns: 150px 1fr;
            gap: 10px;
            margin: 10px 0;
        }

        .info-label {
            font-weight: bold;
            color: #2c3e50;
        }

        .info-value {
            color: #555;
        }

        .relationship-link {
            display: inline-block;
            padding: 5px 10px;
            background: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 5px 5px 5px 0;
        }

        .relationship-link:hover {
            background: #2980b9;
        }

        .media-gallery {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 15px;
        }

        .media-item {
            border: 1px solid #ddd;
            border-radius: 4px;
            overflow: hidden;
            background: white;
        }

        .media-item img {
            width: 100%%;
            height: 200px;
            object-fit: cover;
            cursor: pointer;
        }

        .media-item img:hover {
            opacity: 0.9;
        }

        .media-info {
            padding: 10px;
        }

        .media-title {
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 5px;
        }

        .media-description {
            font-size: 0.9em;
            color: #666;
        }

        .badge {
            display: inline-block;
            padding: 4px 8px;
            background: #3498db;
            color: white;
            border-radius: 4px;
            font-size: 0.85em;
            margin-right: 5px;
        }

        .badge.living {
            background: #27ae60;
        }

        .badge.record {
            background: #f39c12;
        }

        .sources-section {
            margin-top: 30px;
        }

        .sources-list {
            list-style: none;
            padding: 0;
            margin: 15px 0;
        }

        .source-item {
            display: flex;
            align-items: flex-start;
            padding: 15px;
            margin-bottom: 10px;
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 6px;
            transition: box-shadow 0.2s;
        }

        .source-item:hover {
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .source-icon {
            width: 48px;
            height: 48px;
            min-width: 48px;
            background: #f0f0f0;
            border-radius: 4px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 15px;
            font-size: 24px;
        }

        .source-icon.document {
            background: #fff4e6;
            color: #f39c12;
        }

        .source-icon.tree {
            background: #e8f5e9;
            color: #4caf50;
        }

        .source-content {
            flex: 1;
        }

        .source-title {
            font-weight: 600;
            color: #2c3e50;
            margin-bottom: 5px;
            font-size: 1.05em;
        }

        .source-meta {
            font-size: 0.9em;
            color: #666;
            margin-bottom: 8px;
        }

        .source-thumbnail {
            max-width: 150px;
            margin-left: 15px;
            cursor: pointer;
            border-radius: 4px;
            border: 2px solid #e0e0e0;
            transition: transform 0.2s;
        }

        .source-thumbnail:hover {
            transform: scale(1.05);
            border-color: #f39c12;
        }

        .source-count {
            font-size: 0.9em;
            color: #666;
            margin-top: 5px;
        }

        .two-column-container {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin: 30px 0;
        }

        @media (max-width: 900px) {
            .two-column-container {
                grid-template-columns: 1fr;
            }
        }

        .event-list {
            list-style: none;
        }

        .event-item {
            padding: 10px;
            margin: 5px 0;
            background: white;
            border-left: 3px solid #3498db;
        }

        /* Lightbox for images */
        .lightbox {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%%;
            height: 100%%;
            background: rgba(0,0,0,0.9);
            z-index: 1000;
            justify-content: center;
            align-items: center;
        }

        .lightbox.active {
            display: flex;
        }

        .lightbox img {
            max-width: 90%%;
            max-height: 90%%;
            object-fit: contain;
        }

        .lightbox-close {
            position: absolute;
            top: 20px;
            right: 40px;
            color: white;
            font-size: 40px;
            cursor: pointer;
        }

        .lightbox-metadata {
            position: absolute;
            bottom: 40px;
            left: 50%%;
            transform: translateX(-50%%);
            background: rgba(0, 0, 0, 0.8);
            color: white;
            padding: 15px 30px;
            border-radius: 8px;
            font-size: 1.1em;
            max-width: 80%%;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="index.html" class="back-button">← Back to Family Tree</a>

        <h1 id="person-name">Loading...</h1>

        <div class="section" id="basic-info"></div>
        <div class="section" id="relationships"></div>
        <div class="two-column-container">
            <div class="section" id="events"></div>
            <div class="section sources-section" id="sources"></div>
        </div>
        <div class="section" id="media"></div>
    </div>

    <div class="lightbox" id="lightbox">
        <span class="lightbox-close" onclick="closeLightbox()">&times;</span>
        <img id="lightbox-img" src="" alt="Photo">
        <div class="lightbox-metadata" id="lightbox-metadata" style="display: none;"></div>
    </div>

    <script>
        const allPeople = %s;
        const metadata = %s;

        // Get person ID from URL parameter
        const urlParams = new URLSearchParams(window.location.search);
        const personId = urlParams.get('id');

        // Find person in allPeople array
        const person = allPeople.find(p => p.personId === personId);

        if (!person) {
            document.body.innerHTML = '<div style="text-align: center; padding: 40px;"><h1>Person not found</h1><a href="index.html">Back to Family Tree</a></div>';
            throw new Error('Person not found');
        }

        // Update page title
        document.title = (person.fullName || 'Unknown') + ' - Family Tree';

        // Display person info
        document.getElementById('person-name').textContent = person.fullName || 'Unknown';

        // Basic info
        let basicHTML = '<h2>Basic Information</h2><div class="info-grid">';
        if (person.givenName) {
            basicHTML += '<div class="info-label">Given Name:</div><div class="info-value">' + person.givenName + '</div>';
        }
        if (person.surname) {
            basicHTML += '<div class="info-label">Surname:</div><div class="info-value">' + person.surname + '</div>';
        }
        if (person.gender) {
            basicHTML += '<div class="info-label">Gender:</div><div class="info-value">' + person.gender + '</div>';
        }
        basicHTML += '</div>';
        if (person.isLiving) {
            basicHTML += '<span class="badge living">Living</span>';
        }
        document.getElementById('basic-info').innerHTML = basicHTML;

        // Relationships
        let relsHTML = '<h2>Relationships</h2>';
        if (person.parents && person.parents.length > 0) {
            relsHTML += '<div style="margin: 15px 0;"><strong>Parents:</strong><br>';
            person.parents.forEach(p => {
                relsHTML += '<a href="person.html?id=' + encodeURIComponent(p.personId) + '" class="relationship-link">' + p.name + '</a>';
            });
            relsHTML += '</div>';
        }
        if (person.spouses && person.spouses.length > 0) {
            relsHTML += '<div style="margin: 15px 0;"><strong>Spouse(s):</strong><br>';
            person.spouses.forEach(s => {
                relsHTML += '<a href="person.html?id=' + encodeURIComponent(s.personId) + '" class="relationship-link">' + s.name + '</a>';
            });
            relsHTML += '</div>';
        }
        if (person.children && person.children.length > 0) {
            relsHTML += '<div style="margin: 15px 0;"><strong>Children:</strong><br>';
            person.children.forEach(c => {
                relsHTML += '<a href="person.html?id=' + encodeURIComponent(c.personId) + '" class="relationship-link">' + c.name + '</a>';
            });
            relsHTML += '</div>';
        }
        document.getElementById('relationships').innerHTML = relsHTML;

        // Events
        if (person.events && person.events.length > 0) {
            // Build maps of related persons' life events for inference
            let eventDateMap = {}; // date -> {type, relationship, name}

            // Map children's births and deaths
            if (person.children && person.children.length > 0) {
                person.children.forEach(child => {
                    let childPerson = allPeople.find(p => p.personId === child.personId);
                    if (childPerson && childPerson.events) {
                        // Determine gender-specific label
                        let genderLabel = 'child';
                        if (childPerson.gender === 'm') genderLabel = 'son';
                        else if (childPerson.gender === 'f') genderLabel = 'daughter';

                        let birthEvent = childPerson.events.find(e => e.type === 'Birth');
                        if (birthEvent && birthEvent.date) {
                            eventDateMap[birthEvent.date] = {
                                type: 'Birth of ' + genderLabel + ' ' + child.name,
                                name: child.name
                            };
                        }

                        let deathEvent = childPerson.events.find(e => e.type === 'Death');
                        if (deathEvent && deathEvent.date) {
                            eventDateMap[deathEvent.date] = {
                                type: 'Death of ' + genderLabel + ' ' + child.name,
                                name: child.name
                            };
                        }
                    }
                });
            }

            // Map parents' deaths
            if (person.parents && person.parents.length > 0) {
                person.parents.forEach(parent => {
                    let parentPerson = allPeople.find(p => p.personId === parent.personId);
                    if (parentPerson && parentPerson.events) {
                        let genderLabel = 'parent';
                        if (parentPerson.gender === 'm') genderLabel = 'father';
                        else if (parentPerson.gender === 'f') genderLabel = 'mother';

                        let deathEvent = parentPerson.events.find(e => e.type === 'Death');
                        if (deathEvent && deathEvent.date) {
                            eventDateMap[deathEvent.date] = {
                                type: 'Death of ' + genderLabel + ' ' + parent.name,
                                name: parent.name
                            };
                        }
                    }
                });
            }

            // Map spouses' deaths
            if (person.spouses && person.spouses.length > 0) {
                person.spouses.forEach(spouse => {
                    let spousePerson = allPeople.find(p => p.personId === spouse.personId);
                    if (spousePerson && spousePerson.events) {
                        let genderLabel = 'spouse';
                        if (spousePerson.gender === 'm') genderLabel = 'husband';
                        else if (spousePerson.gender === 'f') genderLabel = 'wife';

                        let deathEvent = spousePerson.events.find(e => e.type === 'Death');
                        if (deathEvent && deathEvent.date) {
                            eventDateMap[deathEvent.date] = {
                                type: 'Death of ' + genderLabel + ' ' + spouse.name,
                                name: spouse.name
                            };
                        }
                    }
                });
            }

            // Map siblings' births and deaths
            if (person.parents && person.parents.length > 0) {
                // Find siblings by looking for people who share the same parents
                let siblings = allPeople.filter(p => {
                    if (p.personId === person.personId) return false; // Not self
                    if (!p.parents || p.parents.length === 0) return false;

                    // Check if they share at least one parent
                    return p.parents.some(pParent =>
                        person.parents.some(myParent =>
                            pParent.personId === myParent.personId
                        )
                    );
                });

                siblings.forEach(sibling => {
                    if (sibling.events) {
                        let genderLabel = 'sibling';
                        if (sibling.gender === 'm') genderLabel = 'brother';
                        else if (sibling.gender === 'f') genderLabel = 'sister';

                        let birthEvent = sibling.events.find(e => e.type === 'Birth');
                        if (birthEvent && birthEvent.date) {
                            eventDateMap[birthEvent.date] = {
                                type: 'Birth of ' + genderLabel + ' ' + sibling.fullName,
                                name: sibling.fullName
                            };
                        }

                        let deathEvent = sibling.events.find(e => e.type === 'Death');
                        if (deathEvent && deathEvent.date) {
                            eventDateMap[deathEvent.date] = {
                                type: 'Death of ' + genderLabel + ' ' + sibling.fullName,
                                name: sibling.fullName
                            };
                        }
                    }
                });
            }

            // Helper function to find media associated with an event
            function findEventMedia(event, eventType) {
                if (!person.media || person.media.length === 0) return [];

                let matches = [];
                let eventYear = null;
                if (event.date) {
                    let yearMatch = event.date.toString().match(/\d{4}/);
                    if (yearMatch) eventYear = yearMatch[0];
                }

                person.media.forEach(mediaItem => {
                    let score = 0;

                    // Match by year
                    if (eventYear && mediaItem.date) {
                        if (mediaItem.date.toString().includes(eventYear)) {
                            score += 10;
                        }
                    }

                    // Match by event type in title
                    if (mediaItem.title && eventType) {
                        let titleLower = mediaItem.title.toLowerCase();
                        let typeLower = eventType.toLowerCase();

                        if (titleLower.includes(typeLower) ||
                            (typeLower === 'birth' && titleLower.includes('birth')) ||
                            (typeLower === 'death' && titleLower.includes('death')) ||
                            (typeLower === 'marriage' && (titleLower.includes('marriage') || titleLower.includes('casamento'))) ||
                            (typeLower === 'baptism' && (titleLower.includes('baptism') || titleLower.includes('batismo')))) {
                            score += 20;
                        }
                    }

                    if (score >= 10) {
                        matches.push({media: mediaItem, score: score});
                    }
                });

                // Sort by score and return
                matches.sort((a, b) => b.score - a.score);
                return matches.map(m => m.media);
            }

            // Helper function to extract year from date for sorting
            function extractYear(dateStr) {
                if (!dateStr) return 9999; // Put events with no date at end
                let match = dateStr.toString().match(/\d{4}/);
                return match ? parseInt(match[0]) : 9999;
            }

            // Sort events chronologically by date
            let sortedEvents = [...person.events].sort((a, b) => {
                return extractYear(a.date) - extractYear(b.date);
            });

            let eventsHTML = '<h2>Life Events</h2><ul class="event-list">';
            sortedEvents.forEach(event => {
                // Skip metadata events that aren't real life events
                if (event.type === 'Name' || event.type === 'Gender') {
                    return;
                }

                eventsHTML += '<li class="event-item">';

                // Infer event type if empty
                let eventType = event.type;
                if (!eventType || eventType === '') {
                    // Check if this matches a related person's life event
                    if (event.date && eventDateMap[event.date]) {
                        eventType = eventDateMap[event.date].type;
                    } else {
                        eventType = 'Life Event';
                    }
                }

                eventsHTML += '<strong>' + eventType + '</strong>';
                if (event.date) {
                    eventsHTML += '<br>Date: ' + formatDate(event.date);
                }
                if (event.place) {
                    eventsHTML += '<br>Place: ' + event.place;
                }
                if (event.description) {
                    eventsHTML += '<br><em>' + event.description + '</em>';
                }

                // Show associated media
                let eventMedia = findEventMedia(event, eventType);
                if (eventMedia.length > 0) {
                    eventsHTML += '<br><span style="color: #3498db; font-size: 0.9em;">' + eventMedia.length + ' media</span>';
                    eventsHTML += '<div style="margin-top: 8px; display: flex; flex-wrap: wrap; gap: 6px;">';
                    eventMedia.forEach(media => {
                        let tooltip = [media.title, media.subcategory].filter(x => x).join(' - ');
                        let metadataText = [media.title, media.date, media.subcategory, media.description].filter(x => x).join(' | ');
                        eventsHTML += '<img src="' + media.filePath + '" alt="' + (tooltip || '') + '" title="' + tooltip + '" onclick=\'event.stopPropagation(); openLightbox("' + media.filePath + '", ' + JSON.stringify(metadataText).replace(/'/g, "&apos;") + ')\' style="width: 50px; height: 50px; object-fit: cover; border-radius: 4px; cursor: pointer; border: 1px solid #ddd;">';
                    });
                    eventsHTML += '</div>';
                }

                eventsHTML += '</li>';
            });
            eventsHTML += '</ul>';
            document.getElementById('events').innerHTML = eventsHTML;
        } else {
            document.getElementById('events').style.display = 'none';
        }

        // Media
        if (person.media && person.media.length > 0) {
            let mediaHTML = '<h2>Media (' + person.media.length + ' items)</h2>';
            mediaHTML += '<div class="media-gallery">';
            person.media.forEach(file => {
                const tooltip = [file.title, file.subcategory].filter(x => x).join(' - ');
                const metadataText = [file.title, file.date, file.subcategory, file.description].filter(x => x).join(' | ');

                mediaHTML += '<div class="media-item">';
                mediaHTML += '<img src="' + file.filePath + '" alt="' + (tooltip || person.fullName) + '" onclick=\'openLightbox("' + file.filePath + '", ' + JSON.stringify(metadataText).replace(/'/g, "&apos;") + ')\'>';
                mediaHTML += '<div class="media-info">';
                if (file.title) {
                    mediaHTML += '<div class="media-title">' + file.title + '</div>';
                }
                if (file.description) {
                    mediaHTML += '<div class="media-description">' + file.description + '</div>';
                }
                if (file.date) {
                    mediaHTML += '<div class="media-description">Date: ' + file.date + '</div>';
                }
                mediaHTML += '</div></div>';
            });
            mediaHTML += '</div>';
            document.getElementById('media').innerHTML = mediaHTML;
        } else {
            document.getElementById('media').style.display = 'none';
        }

        // Sources Section (Census, Vital Records, etc.)
        if (person.recordImages && person.recordImages.length > 0) {
            let sourcesHTML = '<h2>Sources (' + person.recordImages.length + ')</h2>';
            sourcesHTML += '<ul class="sources-list">';

            person.recordImages.forEach(record => {
                const recordMetadata = [record.sourceTitle, 'Citation ID: ' + record.citationId].filter(x => x).join(' | ');

                // Determine icon based on database or source type
                let icon = '📄';
                let iconClass = 'document';
                if (record.sourceTitle && (record.sourceTitle.includes('Census') || record.sourceTitle.includes('census'))) {
                    icon = '👥';
                    iconClass = 'document';
                } else if (record.sourceTitle && (record.sourceTitle.includes('Marriage') || record.sourceTitle.includes('Birth') || record.sourceTitle.includes('Death'))) {
                    icon = '📋';
                    iconClass = 'document';
                } else if (record.sourceTitle && (record.sourceTitle.includes('Tree') || record.sourceTitle.includes('Family'))) {
                    icon = '🌳';
                    iconClass = 'tree';
                }

                sourcesHTML += '<li class="source-item">';
                sourcesHTML += '<div class="source-icon ' + iconClass + '">' + icon + '</div>';
                sourcesHTML += '<div class="source-content">';
                sourcesHTML += '<div class="source-title">' + (record.sourceTitle || 'Unknown Source') + '</div>';

                let metaParts = [];
                if (record.databaseId) {
                    metaParts.push('Database: ' + record.databaseId);
                }
                if (record.recordId) {
                    metaParts.push('Record: ' + record.recordId);
                }
                if (metaParts.length > 0) {
                    sourcesHTML += '<div class="source-meta">' + metaParts.join(' • ') + '</div>';
                }

                sourcesHTML += '</div>';

                // Add thumbnail preview
                sourcesHTML += '<img src="' + record.filePath + '" class="source-thumbnail" alt="' + record.sourceTitle + '" onclick=\'openLightbox("' + record.filePath + '", ' + JSON.stringify(recordMetadata).replace(/'/g, "&apos;") + ')\'>';

                sourcesHTML += '</li>';
            });

            sourcesHTML += '</ul>';
            document.getElementById('sources').innerHTML = sourcesHTML;
        } else {
            document.getElementById('sources').style.display = 'none';
        }

        function formatDate(date) {
            if (typeof date === 'object') {
                return JSON.stringify(date);
            }
            return date;
        }

        function openLightbox(imagePath, metadata = '') {
            document.getElementById('lightbox-img').src = imagePath;
            const metadataEl = document.getElementById('lightbox-metadata');
            if (metadata && metadata.trim()) {
                metadataEl.textContent = metadata;
                metadataEl.style.display = 'block';
            } else {
                metadataEl.style.display = 'none';
            }
            document.getElementById('lightbox').classList.add('active');
        }

        function closeLightbox() {
            document.getElementById('lightbox').classList.remove('active');
        }

        // Close lightbox on click outside or ESC key
        document.getElementById('lightbox').addEventListener('click', (e) => {
            if (e.target.id === 'lightbox') closeLightbox();
        });

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') closeLightbox();
        });
    </script>
</body>
</html>`, peopleJSON, metadataJSON)
}
