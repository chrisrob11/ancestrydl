package commands

import "fmt"

// generateHTMLTemplate creates the HTML viewer with embedded JSON data
func generateHTMLTemplate(peopleJSON, metadataJSON string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Family Tree Viewer</title>
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
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 30px;
        }

        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 2.5em;
        }

        .metadata {
            color: #7f8c8d;
            margin-bottom: 30px;
            padding: 15px;
            background: #ecf0f1;
            border-radius: 4px;
        }

        .search-box {
            width: 100%%;
            padding: 12px 20px;
            margin-bottom: 20px;
            font-size: 16px;
            border: 2px solid #bdc3c7;
            border-radius: 4px;
            outline: none;
        }

        .search-box:focus {
            border-color: #3498db;
        }

        .stats {
            display: flex;
            gap: 20px;
            margin-bottom: 30px;
            flex-wrap: wrap;
        }

        .stat-card {
            flex: 1;
            min-width: 200px;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            border-radius: 8px;
            text-align: center;
        }

        .stat-card h3 {
            font-size: 2em;
            margin-bottom: 5px;
        }

        .stat-card p {
            opacity: 0.9;
        }

        .people-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
        }

        .person-card {
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 20px;
            background: white;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .person-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }

        .person-card h3 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 1.3em;
        }

        .person-id {
            color: #95a5a6;
            font-size: 0.85em;
            margin-bottom: 10px;
        }

        .person-info {
            margin: 10px 0;
            color: #555;
        }

        .person-info strong {
            color: #2c3e50;
        }

        .media-gallery {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
            gap: 10px;
            margin-top: 15px;
        }

        .media-gallery img {
            width: 100%%;
            height: 100px;
            object-fit: cover;
            border-radius: 4px;
            cursor: pointer;
            transition: transform 0.2s;
        }

        .media-gallery img:hover {
            transform: scale(1.05);
        }

        .badge {
            display: inline-block;
            padding: 4px 8px;
            background: #3498db;
            color: white;
            border-radius: 4px;
            font-size: 0.85em;
            margin-top: 5px;
            margin-right: 5px;
        }

        .badge.living {
            background: #27ae60;
        }

        .badge.document {
            background: #e74c3c;
        }

        .badge.photo {
            background: #9b59b6;
        }

        .no-results {
            text-align: center;
            padding: 40px;
            color: #95a5a6;
            font-size: 1.2em;
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
        <h1 id="tree-name">Loading Family Tree...</h1>
        <div class="metadata" id="metadata"></div>

        <div class="stats" id="stats"></div>

        <input type="text" class="search-box" id="search" placeholder="Search by name, gender, or any detail...">

        <div style="margin: 20px 0; display: flex; gap: 10px; align-items: center;">
            <label for="sort-select" style="font-weight: bold;">Sort by:</label>
            <select id="sort-select" style="padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 1em;">
                <option value="name">Name (A-Z)</option>
                <option value="name-desc">Name (Z-A)</option>
                <option value="birth-asc">Birth Year (Oldest First)</option>
                <option value="birth-desc">Birth Year (Newest First)</option>
                <option value="death-asc">Death Year (Oldest First)</option>
                <option value="death-desc">Death Year (Newest First)</option>
            </select>
        </div>

        <div class="people-grid" id="people-grid"></div>
    </div>

    <div class="lightbox" id="lightbox">
        <span class="lightbox-close" onclick="closeLightbox()">&times;</span>
        <img id="lightbox-img" src="" alt="Photo">
        <div class="lightbox-metadata" id="lightbox-metadata" style="display: none;"></div>
    </div>

    <script>
        // Embedded data
        const allPeople = %s;
        const metadata = %s;

        // Initialize display
        displayMetadata(metadata);
        displayStats();
        displayPeople(allPeople);

        function displayMetadata(metadata) {
            document.getElementById('tree-name').textContent = metadata.treeName || 'Family Tree';

            const metadataDiv = document.getElementById('metadata');
            const exportDate = new Date(metadata.exportDate).toLocaleDateString();
            metadataDiv.innerHTML = `+"`"+`
                <strong>Tree ID:</strong> ${metadata.treeId} |
                <strong>Exported:</strong> ${exportDate} |
                <strong>Total People:</strong> ${metadata.personCount}
            `+"`"+`;
        }

        function displayStats() {
            const totalPeople = allPeople.length;
            let withMedia = 0;
            let photoCount = 0;
            let documentCount = 0;

            // Count media from person.media field
            allPeople.forEach(person => {
                if (person.media && person.media.length > 0) {
                    withMedia++;
                    person.media.forEach(file => {
                        if (file.category === 'photo' && file.subcategory !== 'document') {
                            photoCount++;
                        } else {
                            documentCount++;
                        }
                    });
                }
            });

            const statsDiv = document.getElementById('stats');
            statsDiv.innerHTML = `+"`"+`
                <div class="stat-card">
                    <h3>${totalPeople}</h3>
                    <p>Total People</p>
                </div>
                <div class="stat-card">
                    <h3>${withMedia}</h3>
                    <p>People with Media</p>
                </div>
                <div class="stat-card">
                    <h3>${documentCount}</h3>
                    <p>Documents</p>
                </div>
                <div class="stat-card">
                    <h3>${photoCount}</h3>
                    <p>Photos</p>
                </div>
            `+"`"+`;
        }

        function displayPeople(people) {
            const grid = document.getElementById('people-grid');

            if (people.length === 0) {
                grid.innerHTML = '<div class="no-results">No people found</div>';
                return;
            }

            grid.innerHTML = people.map(person => {
                const personId = person.personId || 'unknown';
                const name = person.fullName || 'Unknown';
                const media = person.media; // Media is now embedded in person object

                // Get birth/death info
                let birthInfo = '';
                let deathInfo = '';
                if (person.events) {
                    const birth = person.events.find(e => e.type === 'Birth');
                    const death = person.events.find(e => e.type === 'Death');
                    if (birth) {
                        const parts = [];
                        if (birth.date) parts.push(formatDate(birth.date));
                        if (birth.place) parts.push(birth.place);
                        birthInfo = parts.join(' • ');
                    }
                    if (death) {
                        const parts = [];
                        if (death.date) parts.push(formatDate(death.date));
                        if (death.place) parts.push(death.place);
                        deathInfo = parts.join(' • ');
                    }
                }

                const mediaHTML = media && media.length > 0 ? `+"`"+`
                    <div class="media-gallery">
                        ${media.map(file => {
                            const tooltip = [file.title, file.subcategory].filter(x => x).join(' - ');
                            const metadataText = [file.title, file.date, file.subcategory, file.description].filter(x => x).join(' | ');

                            return `+"`"+`<img src="${file.filePath}" alt="${tooltip || name}" title="${tooltip}" onclick='event.stopPropagation(); openLightbox("${file.filePath}", ${JSON.stringify(metadataText).replace(/'/g, "&apos;")})'>`+"`"+`;
                        }).join('')}
                    </div>
                `+"`"+` : '';

                // Count media types
                let photoCount = 0;
                let documentCount = 0;
                if (media && media.length > 0) {
                    media.forEach(file => {
                        if (file.category === 'photo' && file.subcategory !== 'document') {
                            photoCount++;
                        } else {
                            documentCount++;
                        }
                    });
                }

                // Get relationships (now embedded in person object)
                let relationshipsHTML = '';
                if (person.parents && person.parents.length > 0) {
                    relationshipsHTML += `+"`"+`<div class="person-info"><strong>Parents:</strong> ${person.parents.map(p => p.name).join(', ')}</div>`+"`"+`;
                }
                if (person.spouses && person.spouses.length > 0) {
                    relationshipsHTML += `+"`"+`<div class="person-info"><strong>Spouse:</strong> ${person.spouses.map(s => s.name).join(', ')}</div>`+"`"+`;
                }
                if (person.children && person.children.length > 0) {
                    relationshipsHTML += `+"`"+`<div class="person-info"><strong>Children:</strong> ${person.children.map(c => c.name).join(', ')}</div>`+"`"+`;
                }

                return `+"`"+`
                    <div class="person-card" onclick="window.location='person.html?id=${encodeURIComponent(personId)}'" style="cursor: pointer;">
                        <h3>${name}</h3>
                        ${person.gender ? `+"`"+`<div class="person-info"><strong>Gender:</strong> ${person.gender}</div>`+"`"+` : ''}
                        ${birthInfo ? `+"`"+`<div class="person-info"><strong>Birth:</strong> ${birthInfo}</div>`+"`"+` : ''}
                        ${deathInfo ? `+"`"+`<div class="person-info"><strong>Death:</strong> ${deathInfo}</div>`+"`"+` : ''}
                        ${relationshipsHTML}
                        ${person.isLiving ? '<span class="badge living">Living</span>' : ''}
                        ${photoCount > 0 ? `+"`"+`<span class="badge photo">${photoCount} photo${photoCount > 1 ? 's' : ''}</span>`+"`"+` : ''}
                        ${documentCount > 0 ? `+"`"+`<span class="badge document">${documentCount} document${documentCount > 1 ? 's' : ''}</span>`+"`"+` : ''}
                        ${mediaHTML}
                    </div>
                `+"`"+`;
            }).join('');
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

        // Helper function to extract year from event
        function getEventYear(person, eventType) {
            if (!person.events) return null;
            const event = person.events.find(e => e.type === eventType);
            if (!event || !event.date) return null;
            const match = event.date.toString().match(/\d{4}/);
            return match ? parseInt(match[0]) : null;
        }

        // Sort function
        function sortPeople(people, sortBy) {
            const sorted = [...people];

            switch(sortBy) {
                case 'name':
                    sorted.sort((a, b) => (a.fullName || '').localeCompare(b.fullName || ''));
                    break;
                case 'name-desc':
                    sorted.sort((a, b) => (b.fullName || '').localeCompare(a.fullName || ''));
                    break;
                case 'birth-asc':
                    sorted.sort((a, b) => {
                        const yearA = getEventYear(a, 'Birth') || 9999;
                        const yearB = getEventYear(b, 'Birth') || 9999;
                        return yearA - yearB;
                    });
                    break;
                case 'birth-desc':
                    sorted.sort((a, b) => {
                        const yearA = getEventYear(a, 'Birth') || 0;
                        const yearB = getEventYear(b, 'Birth') || 0;
                        return yearB - yearA;
                    });
                    break;
                case 'death-asc':
                    sorted.sort((a, b) => {
                        const yearA = getEventYear(a, 'Death') || 9999;
                        const yearB = getEventYear(b, 'Death') || 9999;
                        return yearA - yearB;
                    });
                    break;
                case 'death-desc':
                    sorted.sort((a, b) => {
                        const yearA = getEventYear(a, 'Death') || 0;
                        const yearB = getEventYear(b, 'Death') || 0;
                        return yearB - yearA;
                    });
                    break;
            }

            return sorted;
        }

        // Get current filtered and sorted people
        function getCurrentPeople() {
            const searchTerm = document.getElementById('search').value.toLowerCase();
            const sortBy = document.getElementById('sort-select').value;

            let people = allPeople;

            // Apply search filter
            if (searchTerm) {
                people = people.filter(person => {
                    const name = (person.fullName || '').toLowerCase();
                    const gender = (person.gender || '').toLowerCase();
                    const id = (person.personId || '').toLowerCase();

                    return name.includes(searchTerm) ||
                           gender.includes(searchTerm) ||
                           id.includes(searchTerm);
                });
            }

            // Apply sort
            people = sortPeople(people, sortBy);

            return people;
        }

        // Search functionality
        document.getElementById('search').addEventListener('input', (e) => {
            displayPeople(getCurrentPeople());
        });

        // Sort functionality
        document.getElementById('sort-select').addEventListener('change', (e) => {
            displayPeople(getCurrentPeople());
        });

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
