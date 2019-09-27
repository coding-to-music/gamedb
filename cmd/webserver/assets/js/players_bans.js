if ($('#player-bans-page').length > 0) {

    const $playersTable = $('table.table');
    const $country = $('#country');

    $('form').on('submit', function (e) {

        $playersTable.DataTable().draw();
        return false;
    });

    $('#country, #state').on('change', function (e) {

        $playersTable.DataTable().draw();
        toggleStateDropDown();
        return false;
    });

    function toggleStateDropDown() {

        $container = $('#state-container');
        if ($country.val() === 'US') {
            $container.removeClass('d-none');
        } else {
            $container.addClass('d-none');
        }
    }

    toggleStateDropDown();

    const options = {
        "language": {
            "zeroRecords": "No players found <a href='/players/add'>Add a Player</a>",
        },
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[6]);
        },
        "columnDefs": [
            // Rank
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[0];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('font-weight-bold')
                },
                "orderable": false,
            },
            // Flag
            {
                "targets": 1,
                "render": function (data, type, row) {
                    if (row[4]) {
                        return '<img data-lazy="' + row[4] + '" alt="" data-lazy-alt="' + row[5] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[5] + '">';
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Player
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                },
                "orderable": false,
            },
            // Game Bans
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[7].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // VAC Bans
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[8].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Last Ban
            {
                "targets": 5,
                "render": function (data, type, row) {
                    if (row[9] > 0) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[10] + '" data-livestamp="' + row[9] + '"></span>';
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderSequence": ["desc"],
            },
        ]
    };

    const searchFields = [
        $('#search'),
        $('#state'),
        $country,
    ];

    $playersTable.gdbTable({tableOptions: options, searchFields: searchFields});
}
