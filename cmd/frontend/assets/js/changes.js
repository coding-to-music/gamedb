if ($('#changes-page').length > 0) {

    const options = {
        'order': [[1, 'desc']],
        'createdRow': function (row, data, dataIndex) {
            $(row).attr('data-link', data[5]);
        },
        'columnDefs': [
            // Change ID
            {
                'targets': 0,
                'render': function (data, type, row) {
                    return '<a href="' + row[5] + '" class="icon-name"><div class="icon"><img src="/assets/img/no-app-image-square.jpg" alt="' + row[1] + '"></div><div class="name">' + row[6] + '</div></a>';
                },
                'createdCell': function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                    $(td).addClass('img');
                },
                'orderable': false,
            },
            // Date
            {
                'targets': 1,
                'render': function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[2] + '" data-livestamp="' + row[1] + '"></span>';
                },
                'createdCell': function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                'orderable': false,
            },
            // Apps
            {
                'targets': 2,
                'render': function (data, type, row) {

                    let apps = [];
                    if (row[3] !== null) {
                        for (let k in row[3]) {
                            if (row[3].hasOwnProperty(k)) {

                                if (row[3][k].name === '') {
                                    row[3][k].name = 'Unknown App';
                                }

                                apps.push('<a href="' + row[3][k].path + '">' + row[3][k].name + '</a>');
                            }
                        }
                    }

                    return apps.join('<br/>');
                },
                'orderable': false,
            },
            // Packages
            {
                'targets': 3,
                'render': function (data, type, row) {

                    let packages = [];
                    if (row[4] !== null) {
                        for (let k in row[4]) {
                            if (row[4].hasOwnProperty(k)) {

                                if (row[4][k].name === '') {
                                    row[4][k].name = 'Unknown Package';
                                }

                                packages.push('<a href="' + row[4][k].path + '">' + row[4][k].name + '</a>');
                            }
                        }
                    }

                    return packages.join('<br/>');
                },
                'orderable': false,
            },
        ],
    };

    const $table = $('table.table');
    const dt = $table.gdbTable({tableOptions: options});

    websocketListener('changes', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = JSON.parse(e.data);

            // Loop changes in websocket data and add each one
            if (isIterable(data.Data)) {

                data.Data.sort(function (a, b) {
                    return Math.sign(a[0] - b[0]);
                });

                for (const v of data.Data) {
                    addDataTablesRow(options, v, info.length, $table);
                }
            }
        }
    });
}
