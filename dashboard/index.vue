<template>
    <div class="layout">
        <Badge
            @click.native="publish(item)"
            v-for="item in Rooms"
            :key="item.StreamPath"
            :count="item.TsCount"
            class="room"
        >
            <Icon type="ios-folder-open-outline" size="100" />
            <div class="size">{{unitFormat(item.TotalSize)}}</div>
            <div>{{item.StreamPath}}</div>
        </Badge>
        <div v-if="Rooms.length==0" class="empty">
            <Icon type="md-wine" size="50" />没有任何文件夹
        </div>
    </div>
</template>

<script>
export default {
    data() {
        return {
            Rooms: []
        };
    },
    methods: {
        fetchlist() {
            window.ajax.getJSON("/ts/list").then(x => {
                this.Rooms = x;
            });
        },
        publish(item) {
            this.$Modal.confirm({
                title: "提示",
                content: "是否发布该目录",
                onOk() {
                    window.ajax.getJSON(
                        "/ts/publish?streamPath=" + item.StreamPath
                    );
                }
            });
        },
        unitFormat: window.unitFormat
    },
    mounted() {
        this.fetchlist();
    }
};
</script>

<style>
@import url("/iview.css");
.layout {
    padding-bottom: 30px;
    display: flex;
    flex-wrap: wrap;
    padding: 30px;
}
.room {
    text-align: center;
}
.room:hover {
    border: 1px solid gray;
    border-radius: 10%;
    cursor: pointer;
}
.size {
    position: absolute;
    top: 40%;
    left: 0;
    right: 0;
}
</style>