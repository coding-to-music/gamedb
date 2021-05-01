if ($('#donate-page').length > 0) {

    const options = {
        'order': [[1, 'desc']],
        'createdRow': function (row, data, dataIndex) {
            if (data[0] > 0) {
                $(row).attr('data-link', data[2]);
            }
        },
        'columnDefs': [
            // Icon / Player Name
            {
                'targets': 0,
                'render': function (data, type, row) {
                    return '<a href="' + row[2] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>';
                },
                'createdCell': function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                'orderable': false,
            },
            // Donation
            {
                'targets': 1,
                'render': function (data, type, row) {
                    return '$' + row[5].toLocaleString();
                },
                'orderable': false,
            },
        ],
    };

    $('#top-donations').gdbTable({tableOptions: options});
    $('#new-donations').gdbTable({tableOptions: options});
}
