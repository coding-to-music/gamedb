if ($('#coop-page').length > 0) {

    const $ids = $('#ids');
    const $search = $('#search');

    loadCoopSearchTable();
    loadCoopGamesTable();

    function loadCoopSearchTable() {

        const options = {
            "order": [[0, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-player-id', data[1]);
            },
            "language": {
                "zeroRecords": function () {
                    return 'No matching players found on Game DB. If a player is missing, <a href="/players/add?search=' + $search.val() + '">add them here</a>.';
                },
            },
            "columnDefs": [
                // Icon / Player Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[4] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[10] + '</div></a>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Games
                {
                    "targets": 1,
                    "render": function (data, type, row) {

                        if (row[0]) {
                            return row[0].toLocaleString();
                        } else {
                            return $lockIcon;
                        }
                    },
                    "orderable": false,
                },
                // Level
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[6].toLocaleString();
                    },
                    "orderable": false,
                },
                // Action
                {
                    "targets": 3,
                    "render": function (data, type, row) {

                        if (row[8]) {
                            return '<a href="' + row[7] + '" ><i class="fas fa-minus"></i> Remove</a>';
                        } else {
                            return '<a href="' + row[7] + '" ><i class="fas fa-plus"></i> Add</a>';
                        }
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false,
                },
                // Community Link
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        if (row[5]) {
                            return '<a href="' + row[5] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
                // Search Score
                {
                    "targets": 5,
                    "render": function (data, type, row) {
                        return row[9].toLocaleString();
                    },
                    "orderable": false,
                    "visible": false,
                },
            ]
        };

        const dt = $('#search-table').gdbTable({
            tableOptions: options,
            searchFields: [$ids, $search],
        });

        dt.on('draw.dt', function (e, settings) {
            if ($search.val()) {
                $('#search-results').show();
            } else {
                $('#search-results').hide();
            }
        });

        $('#players-table').gdbTable({
            tableOptions: options,
            searchFields: [$ids],
        });
    }

    function loadCoopGamesTable() {

        const options = {
            "order": [[0, 'asc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[7]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[7] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Platforms
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[3];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('platforms');
                    },
                    "orderable": false,
                },
                // Achievements
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    "orderable": false,
                },
                // Co-op Tags
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    "orderable": false,
                },
                // Community Link
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        if (row[6]) {
                            return '<a href="' + row[6] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        $('#games-table').gdbTable({
            tableOptions: options,
            searchFields: [$ids],
        });
    }
}
