const $statPage = $('#stat-page');

if ($statPage.length > 0) {

    loadAjaxOnObserve({
        'stat-chart': statHighChart,
        'games': statTable,
    });

    function statTable() {

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[8]);
                $(row).attr('data-player-id', data[0]);
            },
            "columnDefs": [
                // Icon / App Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Players
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    "orderSequence": ["desc"],
                },
                // Price
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    "orderSequence": ["desc"],
                },
                // Score
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "orderSequence": ["desc"],
                },
                // Link
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        if (row[7]) {
                            return '<a href="' + row[7] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        $('#games').gdbTable({
            tableOptions: options,
            searchFields: [
                $('#items-search'),
            ],
        });
    }

    function statHighChart() {

        $.ajax({
            type: "GET",
            url: '/' + $statPage.attr('data-stat-type') + '/' + $statPage.attr('data-stat-id') + '/time.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (data === null) {
                    data = [];
                }

                const yAxis = {
                    allowDecimals: false,
                    title: {
                        text: ''
                    },
                    labels: {
                        enabled: false
                    },
                };

                Highcharts.chart('stat-chart', $.extend(true, {}, defaultChartOptions, {
                    yAxis: [
                        yAxis,
                        yAxis,
                        yAxis,
                        yAxis,
                        yAxis,
                    ],
                    tooltip: {
                        formatter: function () {

                            const day = moment(this.key).format("dddd DD MMM YYYY");

                            switch (this.series.name) {
                                case 'Apps':
                                    return Math.round(this.y).toLocaleString() + ' games with tag on ' + day;
                                case 'Apps (%)':
                                    return this.y.toLocaleString() + '% of games have tag ' + day;
                                case 'Mean Players':
                                    return this.y.toLocaleString() + ' mean max weakly players on ' + day;
                                case 'Mean Price (' + user.userCurrencySymbol + ')':
                                    return user.userCurrencySymbol + ' ' + (this.y / 100).toFixed(2).toLocaleString() + ' mean price on ' + day;
                                case 'Mean Review Score':
                                    return this.y.toLocaleString() + '% mean review score on ' + day;
                                case 'Median Players':
                                    return this.y.toLocaleString() + ' median max weakly players on ' + day;
                                case 'Median Price (' + user.userCurrencySymbol + ')':
                                    return user.userCurrencySymbol + ' ' + (this.y / 100).toFixed(2).toLocaleString() + ' median price on ' + day;
                                case 'Median Review Score':
                                    return this.y.toLocaleString() + '% median review score on ' + day;
                            }
                        },
                    },
                    series: [
                        {
                            name: 'Apps',
                            data: data['max_apps_count'],
                            marker: {symbol: 'circle'},
                            yAxis: 0,
                        },
                        {
                            name: 'Apps (%)',
                            data: data['max_apps_percent'],
                            marker: {symbol: 'circle'},
                            yAxis: 1,
                        },
                        {
                            name: 'Mean Players',
                            data: data['max_mean_players'],
                            marker: {symbol: 'circle'},
                            yAxis: 2,
                        },
                        {
                            name: 'Median Players',
                            data: data['max_median_players'],
                            marker: {symbol: 'circle'},
                            yAxis: 2,
                            visible: false,
                        },
                        {
                            name: 'Mean Price (' + user.userCurrencySymbol + ')',
                            data: data['max_mean_price_' + user.prodCC],
                            marker: {symbol: 'circle'},
                            yAxis: 3,
                        },
                        {
                            name: 'Median Price (' + user.userCurrencySymbol + ')',
                            data: data['max_median_price_' + user.prodCC],
                            marker: {symbol: 'circle'},
                            yAxis: 3,
                            visible: false,
                        },
                        {
                            name: 'Mean Review Score',
                            data: data['max_mean_score'],
                            marker: {symbol: 'circle'},
                            yAxis: 4,
                        },
                        {
                            name: 'Median Review Score',
                            data: data['max_median_score'],
                            marker: {symbol: 'circle'},
                            yAxis: 4,
                            visible: false,
                        },
                    ],
                }));
            },
        });
    }
}
